package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
type DFReplicaType string
const(
	PS DFReplicaType = "PS"
    Worker DFReplicaType = "Worker"	
)

type DFReplicaSpec struct {
	Replicas *int32              `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
    Template *v1.PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`
}
// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// DfJobSpec defines the desired state of DfJob
type DfJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	ReplicaSpecs map[DFReplicaType]*DFReplicaSpec `json:"dfreplicaSpecs"`
}



// DfJobStatus defines the observed state of DfJob
type DfJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DfJob is the Schema for the dfjobs API
// +k8s:openapi-gen=true
type DfJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DfJobSpec   `json:"spec,omitempty"`
	Status DfJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DfJobList contains a list of DfJob
type DfJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DfJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DfJob{}, &DfJobList{})
}
