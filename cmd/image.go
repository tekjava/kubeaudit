package cmd

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
)

var imgConfig imgFlags

type imgFlags struct {
	img  string
	tag  string
	name string
}

func (image *imgFlags) splitImageString() (err error) {
	tokens := strings.Split(image, ":")
	if len(tokens) < 2 {
		err = errors.New("Image does not container name and tag")
		return
	}
	image.name = tokens[0]
	image.tag = tokens[1]
	return
}

func printResultImg(results []Result) {
	for _, result := range results {
		switch result.err {
		case E_INFO:
			log.WithFields(log.Fields{
				"type": result.kubeType,
				"tag":  result.img_tag}).Info(result.namespace,
				"/", result.name)
		case EIMAGE_TAG_MISSING:
			log.WithFields(log.Fields{
				"type": result.kubeType,
				"tag":  result.img_tag}).Error("Image tag was missing ", result.namespace,
				"/", result.name)
		case EINCORRECT_IMAGE:
			log.WithFields(log.Fields{
				"type": result.kubeType,
				"tag":  result.img_tag}).Error("Image tag was incorrect ", result.namespace,
				"/", result.name)
		}
	}
}

func checkImage(container apiv1.Container, image imgFlags, result *Result) {
	contImage := imgFlags{img: container.Image}
	err := contImage.splitImageString()
	// Image name was proper but image tag was missing
	if err != nil && contImage.name == image.name {
		result.err = EIMAGE_TAG_MISSING
		return
	}

	if contImg == compImg && contTag != compTag {
		result.err = EIMAGE_TAG_INCORRECT
		result.img_name = contImage.name
		result.img_tag = contImage.tag
	}
	return
}

func auditImages(image imgFlags, items Items) (results []Result) {
	for _, item := range items.Iter() {
		containers, result := containerIter(item)
		for _, container := range containers {
			checkImage(container, image, result)
			if result != nil && result.err > 0 {
				results = append(results, *result)
				break
			}
		}
	}
	printResultImg(results)
	defer wg.Done()
	return
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Audit container images",
	Long: `This command audits a container against a given image:tag

An INFO log is given when a container has a matching image:tag
An ERROR log is generated when a container does not match the image:tag

This command is also a root command, check kubeaudit sc --help

Example usage:
kubeaudit image --image gcr.io/google_containers/echoserver:1.7
kubeaudit image -i gcr.io/google_containers/echoserver:1.7`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(imgConfig.img) == 0 {
			log.Error("Empty image name. Are you missing the image flag?")
			return
		}
		err := imgConfig.splitImageString()
		if err != nil {
			log.Error(err)
			return
		}

		if rootConfig.json {
			log.SetFormatter(&log.JSONFormatter{})
		}

		if rootConfig.manifest != "" {
			wg.Add(1)
			resource := getKubeResource(rootConfig.manifest)
			auditSecurityContext(resource)
			wg.Wait()
		} else {
			kube, err := kubeClient(rootConfig.kubeConfig)
			if err != nil {
				log.Error(err)
			}

			// fetch deployments, statefulsets, daemonsets
			// and pods which do not belong to another abstraction
			deployments := getDeployments(kube)
			statefulSets := getStatefulSets(kube)
			daemonSets := getDaemonSets(kube)
			replicationControllers := getReplicationControllers(kube)
			pods := getPods(kube)

			wg.Add(5)
			go auditImages(imgConfig, kubeAuditStatefulSets{list: statefulSets})
			go auditImages(imgConfig, kubeAuditDaemonSets{list: daemonSets})
			go auditImages(imgConfig, kubeAuditPods{list: pods})
			go auditImages(imgConfig, kubeAuditReplicationControllers{list: replicationControllers})
			go auditImages(imgConfig, kubeAuditDeployments{list: deployments})
			wg.Wait()
		}
	},
}

func init() {
	RootCmd.AddCommand(imageCmd)
	imageCmd.Flags().StringVarP(&imgConfig.img, "image", "i", "", "image to check against")
}
