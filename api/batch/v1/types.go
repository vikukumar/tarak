// Package v1 contains API types for the batch API group.
package v1

import (
	appsv1 "github.com/vikukumar/tarak/api/apps/v1"
	"github.com/vikukumar/tarak/api/meta"
)

// ─── Job ─────────────────────────────────────────────────────────────────────

// Job represents the configuration of a single job.
type Job struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            JobSpec   `json:"spec,omitempty"`
	Status          JobStatus `json:"status,omitempty"`
}

// JobSpec describes how the job execution will look like.
type JobSpec struct {
	// Parallelism specifies the maximum desired number of pods the job should run at any given time.
	Parallelism *int32 `json:"parallelism,omitempty"`
	// Completions specifies the desired number of successfully finished pods the job should be run with.
	Completions *int32 `json:"completions,omitempty"`
	// ActiveDeadlineSeconds specifies the duration in seconds relative to the startTime
	// that the job may be continuously active before the system tries to terminate it.
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
	// PodFailurePolicy specifies how the Job treats failed pods.
	PodFailurePolicy *PodFailurePolicy `json:"podFailurePolicy,omitempty"`
	// SuccessPolicy specifies the policy when the Job can be declared as succeeded.
	SuccessPolicy *SuccessPolicy `json:"successPolicy,omitempty"`
	// BackoffLimit specifies the number of retries before marking this job failed.
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`
	// BackoffLimitPerIndex specifies the maximum number of retries per index before marking that index as failed.
	BackoffLimitPerIndex *int32 `json:"backoffLimitPerIndex,omitempty"`
	// MaxFailedIndexes specifies the maximal number of failed indexes before marking the Job as failed.
	MaxFailedIndexes *int32 `json:"maxFailedIndexes,omitempty"`
	// Selector is a label query over pods that should match the pod count.
	Selector *meta.LabelSelector `json:"selector,omitempty"`
	// ManualSelector controls generation of pod labels and pod selectors.
	ManualSelector *bool `json:"manualSelector,omitempty"`
	// Template is the object that describes the pod that will be created when executing a job.
	Template appsv1.PodTemplateSpec `json:"template"`
	// TTLSecondsAfterFinished limits the lifetime of a Job that has finished execution.
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
	// CompletionMode specifies how Pod completions are tracked.
	CompletionMode *CompletionMode `json:"completionMode,omitempty"`
	// Suspend specifies whether the Job controller should create Pods or not.
	Suspend *bool `json:"suspend,omitempty"`
	// PodReplacementPolicy specifies when to create replacement Pods.
	PodReplacementPolicy *PodReplacementPolicy `json:"podReplacementPolicy,omitempty"`
	// ManagedBy field indicates the controller that manages a Job.
	ManagedBy *string `json:"managedBy,omitempty"`
}

type CompletionMode string

const (
	NonIndexedCompletion CompletionMode = "NonIndexed"
	IndexedCompletion    CompletionMode = "Indexed"
)

type PodReplacementPolicy string

const (
	TerminatingOrFailed PodReplacementPolicy = "TerminatingOrFailed"
	Failed              PodReplacementPolicy = "Failed"
)

// PodFailurePolicy describes how failed pods influence the backoffLimit.
type PodFailurePolicy struct {
	Rules []PodFailurePolicyRule `json:"rules"`
}

// PodFailurePolicyRule describes how a pod failure is handled when the requirements are met.
type PodFailurePolicyRule struct {
	Action          PodFailurePolicyAction                   `json:"action"`
	OnExitCodes     *PodFailurePolicyOnExitCodesRequirement  `json:"onExitCodes,omitempty"`
	OnPodConditions []PodFailurePolicyOnPodConditionsPattern `json:"onPodConditions,omitempty"`
}

type PodFailurePolicyAction string

const (
	PodFailurePolicyActionFailJob   PodFailurePolicyAction = "FailJob"
	PodFailurePolicyActionFailIndex PodFailurePolicyAction = "FailIndex"
	PodFailurePolicyActionIgnore    PodFailurePolicyAction = "Ignore"
	PodFailurePolicyActionCount     PodFailurePolicyAction = "Count"
)

// PodFailurePolicyOnExitCodesRequirement describes the requirement for handling a failed pod based on its exit codes.
type PodFailurePolicyOnExitCodesRequirement struct {
	ContainerName *string                             `json:"containerName,omitempty"`
	Operator      PodFailurePolicyOnExitCodesOperator `json:"operator"`
	Values        []int32                             `json:"values"`
}

type PodFailurePolicyOnExitCodesOperator string

const (
	PodFailurePolicyOnExitCodesOpIn    PodFailurePolicyOnExitCodesOperator = "In"
	PodFailurePolicyOnExitCodesOpNotIn PodFailurePolicyOnExitCodesOperator = "NotIn"
)

// PodFailurePolicyOnPodConditionsPattern describes a pattern for matching an actual pod condition type.
type PodFailurePolicyOnPodConditionsPattern struct {
	Type   string               `json:"type"`
	Status meta.ConditionStatus `json:"status"`
}

// SuccessPolicy describes when a Job can be declared as succeeded based on the success of some indexes.
type SuccessPolicy struct {
	Rules []SuccessPolicyRule `json:"rules"`
}

// SuccessPolicyRule describes rule for declaring a Job as succeeded.
type SuccessPolicyRule struct {
	SucceededIndexes *string `json:"succeededIndexes,omitempty"`
	SucceededCount   *int32  `json:"succeededCount,omitempty"`
}

// JobStatus represents the current state of a Job.
type JobStatus struct {
	// Conditions represent the latest available observations of an object's current state.
	Conditions []JobCondition `json:"conditions,omitempty"`
	// UncountedTerminatedPods holds UIDs of Pods that have terminated but haven't been accounted in the Job status.
	UncountedTerminatedPods *UncountedTerminatedPods `json:"uncountedTerminatedPods,omitempty"`
	// StartTime represents time when the job was acknowledged by the Job controller.
	StartTime *meta.Time `json:"startTime,omitempty"`
	// CompletionTime represents the time when the job was completed.
	CompletionTime *meta.Time `json:"completionTime,omitempty"`
	// Active is the number of pending and running pods which are not terminating.
	Active int32 `json:"active,omitempty"`
	// Succeeded is the number of pods which reached phase Succeeded.
	Succeeded int32 `json:"succeeded,omitempty"`
	// Failed is the number of pods which reached phase Failed.
	Failed int32 `json:"failed,omitempty"`
	// Terminating is the number of pods which are terminating.
	Terminating *int32 `json:"terminating,omitempty"`
	// CompletedIndexes holds the completed indexes when .spec.completionMode = "Indexed" in a text format.
	CompletedIndexes string `json:"completedIndexes,omitempty"`
	// FailedIndexes holds the failed indexes when .spec.backoffLimitPerIndex is set.
	FailedIndexes *string `json:"failedIndexes,omitempty"`
	// Ready is the number of pods where the Readiness Gate is set.
	Ready *int32 `json:"ready,omitempty"`
}

// UncountedTerminatedPods holds UIDs of Pods that have terminated but haven't been accounted in Job status counters.
type UncountedTerminatedPods struct {
	Succeeded []string `json:"succeeded,omitempty"`
	Failed    []string `json:"failed,omitempty"`
}

// JobCondition describes current state of a job.
type JobCondition struct {
	Type               JobConditionType     `json:"type"`
	Status             meta.ConditionStatus `json:"status"`
	LastProbeTime      meta.Time            `json:"lastProbeTime,omitempty"`
	LastTransitionTime meta.Time            `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

type JobConditionType string

const (
	JobSuspended          JobConditionType = "Suspended"
	JobComplete           JobConditionType = "Complete"
	JobFailed             JobConditionType = "Failed"
	JobFailureTarget      JobConditionType = "FailureTarget"
	JobSuccessCriteriaMet JobConditionType = "SuccessCriteriaMet"
)

// JobList is a collection of jobs.
type JobList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Job `json:"items"`
}

// ─── CronJob ─────────────────────────────────────────────────────────────────

// CronJob represents the configuration of a single cron job.
type CronJob struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            CronJobSpec   `json:"spec,omitempty"`
	Status          CronJobStatus `json:"status,omitempty"`
}

// CronJobSpec describes how the job execution will look like and when it will actually run.
type CronJobSpec struct {
	// Schedule is the Cron format schedule string, following standard cron syntax.
	Schedule string `json:"schedule"`
	// TimeZone is the name of the timezone used for interpreting the schedule.
	TimeZone *string `json:"timeZone,omitempty"`
	// StartingDeadlineSeconds is the deadline in seconds for starting the job if it misses its scheduled time.
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`
	// ConcurrencyPolicy specifies how to treat concurrent executions of a Job.
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`
	// Suspend this flag tells the controller to suspend subsequent executions.
	Suspend *bool `json:"suspend,omitempty"`
	// JobTemplate is the object that describes the job that will be created when executing a CronJob.
	JobTemplate JobTemplateSpec `json:"jobTemplate"`
	// SuccessfulJobsHistoryLimit is the number of successful finished jobs to retain.
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`
	// FailedJobsHistoryLimit is the number of failed finished jobs to retain.
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

type ConcurrencyPolicy string

const (
	AllowConcurrent   ConcurrencyPolicy = "Allow"
	ForbidConcurrent  ConcurrencyPolicy = "Forbid"
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// JobTemplateSpec describes the data a Job should have when created from a template.
type JobTemplateSpec struct {
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            JobSpec `json:"spec,omitempty"`
}

// CronJobStatus represents the current state of a cron job.
type CronJobStatus struct {
	// Active is a list of pointers to currently running jobs.
	Active []meta.ObjectMeta `json:"active,omitempty"`
	// LastScheduleTime information when was the last time the job was successfully scheduled.
	LastScheduleTime *meta.Time `json:"lastScheduleTime,omitempty"`
	// LastSuccessfulTime information when was the last time the job successfully completed.
	LastSuccessfulTime *meta.Time `json:"lastSuccessfulTime,omitempty"`
}

// CronJobList is a collection of cron jobs.
type CronJobList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []CronJob `json:"items"`
}
