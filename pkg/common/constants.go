package common

const (
	// AideConfigLabelKey tells us if a specific ConfigMap is an AIDE config
	AideConfigLabelKey = "file-integrity.openshift.io/aide-conf"
	// AideConfigUpdatedAnnotationKey tells us if an aide config needs updating
	AideConfigUpdatedAnnotationKey = "file-integrity.openshift.io/updated"
	// AideDatabaseReinitAnnotationKey tells us if an aide config needs updating
	AideDatabaseReinitAnnotationKey = "file-integrity.openshift.io/re-init"
	// IntegrityLogLabelKey tells us that a log was generated by the log collector
	IntegrityLogLabelKey = "file-integrity.openshift.io/log"
	// IntegrityLogResultLabelKey tells us that the configMap represents a result log (a log we decided to keep)
	IntegrityLogResultLabelKey = "file-integrity.openshift.io/result-log"
	// IntegrityConfigMapOwnerLabelKey tells us what FileIntegrity object owns a specific ConfigMap
	IntegrityConfigMapOwnerLabelKey = "file-integrity.openshift.io/owner"
	// IntegrityConfigMapNodeLabelKey tells us from which node did the configmap come from
	IntegrityConfigMapNodeLabelKey = "file-integrity.openshift.io/node"
	// IntegrityLogContentKey is the key in the configmap where the logs are stored
	IntegrityLogContentKey = "integritylog"
	// IntegrityLogErrorAnnotationKey indicates that there was an error in the logcollector
	IntegrityLogErrorAnnotationKey = "file-integrity.openshift.io/log-errormsg"
	// CompressedLogsIndicatorLabelKey indicates the log has been compressed
	CompressedLogsIndicatorLabelKey    = "file-integrity.openshift.io/compressed"
	IntegrityLogFilesAddedAnnotation   = "file-integrity.openshift.io/files-added"
	IntegrityLogFilesRemovedAnnotation = "file-integrity.openshift.io/files-removed"
	IntegrityLogFilesChangedAnnotation = "file-integrity.openshift.io/files-changed"
	AideInitScriptConfigMapName        = "aide-init"
	AideReinitScriptConfigMapName      = "aide-reinit"
	AideScriptConfigMapName            = "aide-script"
	AideScriptPath                     = "/scripts/aide.sh"
	DaemonSetPrefix                    = "aide-ds"
	DefaultConfDataKey                 = "aide.conf"
	FileIntegrityNamespace             = "openshift-file-integrity"
	OperatorServiceAccountName         = "file-integrity-operator"
	ReinitDaemonSetPrefix              = "aide-reinit-ds"
)
