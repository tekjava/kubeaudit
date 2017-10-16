package cmd

const (
	KubeAuditInfo = iota
	ErrorImageTagMissing
	ErrorImageTagIncorrect
	ErrorSecurityContextNIL
	ErrorReadOnlyRootFilesystemNIL
	ErrorReadOnlyRootFilesystemFalse
	ErrorServiceAccountTokenDeprecated
	ErrorServiceAccountTokenNoName
	ErrorServiceAccountTokenNIL
)
