package dfjob

import (
"context"
"strconv"
// "strings"
corev1 "k8s.io/api/core/v1"
dlflowv1alpha1 "github.com/df-operator/pkg/apis/cache/v1alpha1"
"k8s.io/apimachinery/pkg/api/errors"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
"k8s.io/apimachinery/pkg/runtime"
"k8s.io/apimachinery/pkg/types"
"sigs.k8s.io/controller-runtime/pkg/client"
"sigs.k8s.io/controller-runtime/pkg/controller"
"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
"sigs.k8s.io/controller-runtime/pkg/handler"
"sigs.k8s.io/controller-runtime/pkg/manager"
"sigs.k8s.io/controller-runtime/pkg/reconcile"
logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
"sigs.k8s.io/controller-runtime/pkg/source"
"k8s.io/apimachinery/pkg/util/intstr"
// "k8s.io/client-go/tools/clientcmd"
// "k8s.io/client-go/kubernetes"
)

var log = logf.Log.WithName("controller_dfjob")
//
/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DfJob Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDfJob{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dfjob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DfJob
	err = c.Watch(&source.Kind{Type: &dlflowv1alpha1.DfJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner DfJob
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &dlflowv1alpha1.DfJob{},
		})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDfJob{}
var p int 
var w int 
var index int
var bestWorkerNum int 
var restGpuNum int 
var psnum int 
var workernum int 
// ReconcileDfJob reconciles a DfJob object
type ReconcileDfJob struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DfJob object and makes changes based on the state read
// and what is in the DfJob.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDfJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling DfJob")
	// Fetch the DfJob instance
	dfjob := &dlflowv1alpha1.DfJob{}
	err := r.client.Get(context.TODO(), request.NamespacedName, dfjob)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	restGpuNum = 0 
	bestWorkerNum = 6
	index = 0
	p = 1
	w = 0
	opts := &client.ListOptions{}
	nodeList:=&corev1.NodeList{}
	err=r.client.List(context.TODO(),opts,nodeList)
	if err!=nil{
		return reconcile.Result{}, err		
	}
	podList:=&corev1.PodList{}
	err=r.client.List(context.TODO(),opts,podList)
	if err!=nil{
		return reconcile.Result{},err
	}
	resourceUsage := make(map[string]*ResourceUsage)
	for _, node := range nodeList.Items {
		resourceUsage[node.Name] = &ResourceUsage{}
	}				
	for _, p := range podList.Items {
		if p.Status.Phase=="Running"{			
			if p.Spec.NodeName == "" {
				continue
			}
			for _, c := range p.Spec.Containers {
	        	// message := fmt.Sprintf("info!!!!!!!! (%s): ", c.Resources.Requests)
	        	// log.Println(message)
				gpuinfo := c.Resources.Requests["nvidia.com/gpu"]
				gpuString := gpuinfo.String()
				gpus, err := strconv.Atoi(gpuString)
				if err != nil {
					// return nil, err
					log.Info("Node Gpu allocated number error!")
				}
				ru := resourceUsage[p.Spec.NodeName]
				ru.GPU += gpus
			}
		}
	}
	for _,node:=range nodeList.Items{		
		nodeinfo:=node.Status.Allocatable["nvidia.com/gpu"]	
		//core_v1找一下allocated_resource字段。
		nodeAllocatableGpuNumString:=nodeinfo.String()
		nodeAllocatableGpuNum,err:=strconv.Atoi(nodeAllocatableGpuNumString)
		if err!=nil{
			log.Info("node Gpu Num stringtoint error!!")				
			return reconcile.Result{}, err
		}
		nodeRestGpuNum:=nodeAllocatableGpuNum - resourceUsage[node.Name].GPU
		//only gpu memory and cpu?
		//所有pod 的request+起来。			
		restGpuNum=restGpuNum+nodeRestGpuNum
		//不考虑具体的placement，如果要考虑ps跟worker的具体placement,应该放在scheduler里面实现。
		//或者说直接在operator里面指定ps worker的node_selector标签，这样的话就不需要实现scheduler。
	}
	//dfjob crd has been found
	if restGpuNum>bestWorkerNum{
		workernum=bestWorkerNum
		psnum=bestWorkerNum+4
	}else{
		workernum=restGpuNum
		psnum=restGpuNum+4
	}
	if dfjob.Spec.ReplicaSpecs["PS"].Replicas!=nil{
		if *dfjob.Spec.ReplicaSpecs["PS"].Replicas!=0{
			psnum=int(*dfjob.Spec.ReplicaSpecs["PS"].Replicas)
		}
	}
	if dfjob.Spec.ReplicaSpecs["Worker"].Replicas!=nil{
		if *dfjob.Spec.ReplicaSpecs["Worker"].Replicas!=0{
			workernum=int(*dfjob.Spec.ReplicaSpecs["Worker"].Replicas)
		}
	}
	reqLogger=log.WithValues("worker number: ",workernum)
	reqLogger.Info("Output the worker num")
	// psnum=0
	// workernum=0
	//create ps nodes and service
	pspod,psservice:=newPsPodForCR(dfjob,index,psnum,workernum)
	index=index+1
	if err := controllerutil.SetControllerReference(dfjob, pspod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err := controllerutil.SetControllerReference(dfjob, psservice, r.scheme); err != nil {
		return reconcile.Result{}, err
	}	
	//	
	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pspod.Name, Namespace: pspod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err=r.client.Create(context.TODO(),pspod)
		if err!=nil{
			reqLogger= log.WithValues("pspod error!!!!!!",1)
			reqLogger.Info("*pspod error!!!!")				
			return reconcile.Result{},err
		}
		err=r.client.Create(context.TODO(),psservice)
		if(err!=nil){
			reqLogger= log.WithValues("psservice error!!!!!!",1)
			reqLogger.Info("*psservice error!!!!")				
			return reconcile.Result{},err
		}
		reqLogger= log.WithValues("w index!!!!!!",w)
		reqLogger.Info("w index!!!!")			
		for index<psnum{
			pspod,psservice:=newPsPodForCR(dfjob,index,psnum,workernum)
			if err := controllerutil.SetControllerReference(dfjob, pspod, r.scheme); err != nil {
				return reconcile.Result{}, err
			}			
			err=r.client.Create(context.TODO(),pspod)
			if err!=nil{
				return reconcile.Result{},err
			}
			if err := controllerutil.SetControllerReference(dfjob, psservice, r.scheme); err != nil {
				return reconcile.Result{}, err
			}			
			err=r.client.Create(context.TODO(),psservice)
			if err!=nil{
				return reconcile.Result{},err
			}								
			index=index+1			
		}
		index=0		
		for index<workernum{		
			workerpod,workerservice:=newWorkerPodForCR(dfjob,index,psnum,workernum)
			if err := controllerutil.SetControllerReference(dfjob, workerpod, r.scheme); err != nil {				
				return reconcile.Result{}, err
			}				
			err=r.client.Create(context.TODO(),workerpod)
			if err!=nil{
				reqLogger.Info("worker create error!!!!")								
				return reconcile.Result{},err
			}
			if err := controllerutil.SetControllerReference(dfjob, workerservice, r.scheme); err != nil {
				return reconcile.Result{}, err
			}				
			err=r.client.Create(context.TODO(),workerservice)
			if(err!=nil){
				return reconcile.Result{},err
			}						
			index=index+1	
		}
	}else if err != nil {
		return reconcile.Result{}, err
	}	
	// Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func newPsPodForCR(cr *dlflowv1alpha1.DfJob,i int,psnum int,workernum int)(*corev1.Pod,*corev1.Service){
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	// pvc := corev1.PersistentVolumeClaimVolumeSource{ClaimName:"ddlp-pv-claim"}
	// volumes = append(volumes, corev1.Volume{Name: "work-dir", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &pvc}})
	volumetype:=corev1.HostPathDirectory	
	// pvc := corev1.PersistentVolumeClaimVolumeSource{ClaimName:"ddlp-pv-claim"}
	hostpath:=corev1.HostPathVolumeSource{Path: "/data/nfs/k8s-tensorflow",Type: &volumetype}
	volumes = append(volumes, corev1.Volume{Name: "work-dir", VolumeSource: corev1.VolumeSource{HostPath: &hostpath}})
    volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "work-dir", MountPath: "/data/tensorflow"})
	str:=strconv.Itoa(i)
	var targetport intstr.IntOrString
	targetport=intstr.FromInt(2222)
	labels:=map[string]string{
		"app":cr.Name+"-ps-"+str,
		"type":"ps",
	}

	psargs:=cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Args
	psargs=append(psargs,"--job_name=ps")
	psargs=append(psargs,"--task_index="+strconv.Itoa(i))
	psspecindex:=len(psargs)
	for iter_ps:=0;iter_ps<psnum;iter_ps++{
		if iter_ps==0{
			psargs=append(psargs,"--ps_hosts="+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222")
		}else{
			psargs[psspecindex]+=","+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222"
		}
	}
	for iter_worker:=0;iter_worker<workernum;iter_worker++{
		if iter_worker==0{
			psargs=append(psargs,"--worker_hosts="+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222")
		}else{
			psargs[psspecindex+1]+=","+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222"
		}
	}
	psenv_value:="{\"environment\": \"cloud\", \"cluster\": {\"ps\":["
	for iter_ps:=0;iter_ps<psnum;iter_ps++{
		psenv_value=psenv_value+"\""+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222"+"\""
		if iter_ps<psnum-1{
			psenv_value=psenv_value+","
		}
	}
	psenv_value=psenv_value+"], \"worker\": ["
	for iter_worker:=0;iter_worker<workernum;iter_worker++{
		psenv_value=psenv_value+"\""+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222"+"\""
		if iter_worker<workernum-1{
			psenv_value=psenv_value+","
		}	
	}
	// task_index:=strings.Split(cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Args[1],"=")[1]
	//psenv_value=psenv_value+"}, \"task\": {\"index\":"+task_index+", \"type\": \"ps\"}}"
	psenv_value=psenv_value+"}, \"task\": {\"index\":"+strconv.Itoa(i)+", \"type\": \"ps\"}}"	
	pspod:=&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name+"-ps-"+str,
			Namespace: cr.Namespace,
			Labels:    labels,
			},
//		Spec:cr.Spec.ReplicaSpecs["PS"].Template.Spec,
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Name,
					Image:           cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Image,
					Command:         cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Command,
					ImagePullPolicy: cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].ImagePullPolicy,
					//Args: []string{"--job_name=ps","--task_index=0","--ps_hosts=example-dfjob-0-ps-1:2222","--worker_hosts=example-dfjob-0-worker-1:2222"},
					Args: psargs,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 2222,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "TF_CONFIG", 
							//Value: "{\"environment\": \"cloud\", \"cluster\": {\"ps\":[\"example-dfjob-0-ps-1:2222\"], \"worker\": [\"example-dfjob-0-worker-1:2222\"]}, \"task\": {\"index\":0, \"type\": \"ps\"}}",
							Value:psenv_value,
						},
						{
							Name:"CUDA_VISIBLE_DEVICES",
							Value:"-1",
						},
					},
					Resources:    cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Resources,
					// VolumeMounts: []corev1.VolumeMount{
					// 	{
					// 		Name:"work-dir",
					// 		MountPath:"/data/tensorflow/",
					// 	},
					// },
					VolumeMounts: volumeMounts,
				},
			},
			RestartPolicy: "Never",
			// Volumes: []corev1.Volume{
			// 	{
			// 		Name:"work-dir",
			// 		VolumeSource:corev1.VolumeSource{
			// 			HostPath:&corev1.HostPathVolumeSource{
   //             				Path: "/data/nfs/k8s-tensorflow",
   //              			Type: &volumetype,							
			// 			},
			// 		},	
			// 	},
			// },
			Volumes:volumes,
			// SchedulerName:"ddlp-scheduler",
		},

	}
	psservice:=&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name+"-ps-"+str,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec:corev1.ServiceSpec{
			Ports:[]corev1.ServicePort{{
					Port: 2222,
					Name: "replica-port",
					Protocol: "TCP",
					TargetPort: targetport,
				},},
			Selector:labels,
			Type:"ClusterIP",
		},


	}
		return pspod,psservice

}
//func deletepspodandservice
func newWorkerPodForCR(cr *dlflowv1alpha1.DfJob,i int,psnum int,workernum int)(*corev1.Pod,*corev1.Service){
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	volumetype:=corev1.HostPathDirectory	
	// pvc := corev1.PersistentVolumeClaimVolumeSource{ClaimName:"ddlp-pv-claim"}
	hostpath:=corev1.HostPathVolumeSource{Path: "/data/nfs/k8s-tensorflow",Type: &volumetype}
	volumes = append(volumes, corev1.Volume{Name: "work-dir", VolumeSource: corev1.VolumeSource{HostPath: &hostpath}})
    volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "work-dir", MountPath: "/data/tensorflow"})
	str:=strconv.Itoa(i)
	var targetport intstr.IntOrString
	targetport=intstr.FromInt(2222)
	labels:=map[string]string{
		"app":cr.Name+"-worker-"+str,
		"type":"worker",
	}	
	workerargs:=cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].Args
	workerargs=append(workerargs,"--job_name=worker")
	workerargs=append(workerargs,"--task_index="+strconv.Itoa(i))
	workerspecindex:=len(workerargs)
	for iter_ps:=0;iter_ps<psnum;iter_ps++{
		if iter_ps==0{
			workerargs=append(workerargs,"--ps_hosts="+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222")
		}else{
			workerargs[workerspecindex]+=","+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222"
		}
	}
	for iter_worker:=0;iter_worker<workernum;iter_worker++{
		if iter_worker==0{
			workerargs=append(workerargs,"--worker_hosts="+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222")
		}else{
			workerargs[workerspecindex+1]+=","+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222"
		}
	}
	workerenv_value:="{\"environment\": \"cloud\", \"cluster\": {\"ps\":["
	for iter_ps:=0;iter_ps<psnum;iter_ps++{
		workerenv_value=workerenv_value+"\""+cr.Name+"-ps-"+strconv.Itoa(iter_ps)+":2222"+"\""
		if iter_ps<psnum-1{
			workerenv_value=workerenv_value+","
		}
	}
	workerenv_value=workerenv_value+"], \"worker\": ["
	for iter_worker:=0;iter_worker<workernum;iter_worker++{
		workerenv_value=workerenv_value+"\""+cr.Name+"-worker-"+strconv.Itoa(iter_worker)+":2222"+"\""
		if iter_worker<workernum-1{
			workerenv_value=workerenv_value+","
		}	
	}
	// task_index:=strings.Split(cr.Spec.ReplicaSpecs["PS"].Template.Spec.Containers[0].Args[1],"=")[1]
	// workerenv_value=workerenv_value+"}, \"task\": {\"index\":"+task_index+", \"type\": \"worker\"}}"
	workerenv_value=workerenv_value+"}, \"task\": {\"index\":"+strconv.Itoa(i)+", \"type\": \"worker\"}}"
	workerpod:=&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name+"-worker-"+str,
			Namespace: cr.Namespace,
			Labels:    labels,
			},
//		Spec:cr.Spec.ReplicaSpecs["Worker"].Template.Spec,
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].Name,
					Image:           cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].Image,
					Command:         cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].Command,
					ImagePullPolicy: cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].ImagePullPolicy,
					//RestartPolicy: cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].RestartPolicy,
					//Args: []string{"--job_name=ps","--task_index=0","--ps_hosts=example-dfjob-0-ps-1:2222","--worker_hosts=example-dfjob-0-worker-1:2222"},
					Args:workerargs,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 2222,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "TF_CONFIG", 
							//Value: "{\"environment\": \"cloud\", \"cluster\": {\"ps\":[\"example-dfjob-0-ps-1:2222\"], \"worker\": [\"example-dfjob-0-worker-1:2222\"]}, \"task\": {\"index\":0, \"type\": \"ps\"}}",
							Value:workerenv_value,
						},
						// {
						// 	Name:"CUDA_VISIBLE_DEVICES",
						// 	Value:"-1",
						// },
					},
					Resources:    cr.Spec.ReplicaSpecs["Worker"].Template.Spec.Containers[0].Resources,
					// VolumeMounts: []corev1.VolumeMount{
					// 	{
					// 		Name:"work-dir",
					// 		MountPath:"/data/tensorflow/",
					// 	},
					// },
					VolumeMounts:volumeMounts,
				},
			},
			RestartPolicy: "Never",

			// Volumes: []corev1.Volume{
			// 	{
			// 		Name:"work-dir",
			// 		VolumeSource:corev1.VolumeSource{
			// 			HostPath:&corev1.HostPathVolumeSource{
   //             				Path: "/data/nfs/k8s-tensorflow",
   //              			Type: &volumetype,							
			// 			},
			// 		},	
			// 	},
			// },
			Volumes:volumes,
				// SchedulerName:"ddlp-scheduler",			
		},
	}
	workerservice:=&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name+"-worker-"+str,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec:corev1.ServiceSpec{
			Ports:[]corev1.ServicePort{{
					Port: 2222,
					Name: "replica-port",
					Protocol: "TCP",
					TargetPort: targetport,
				},},
			Selector:labels,
			Type:"ClusterIP",
		},
	}
		return workerpod,workerservice

}
//func delete workerpod and service
