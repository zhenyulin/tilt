package k8s

import (
	"context"
	"time"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

type InformEvent interface {
	informEvent()
}

type AddInformEvent struct {
	Obj interface{}
}

type UpdateInformEvent struct {
	OldObj interface{}
	NewObj interface{}
}

type DeleteInformEvent struct {
	Last interface{}
}

func (AddInformEvent) informEvent()    {}
func (UpdateInformEvent) informEvent() {}
func (DeleteInformEvent) informEvent() {}

func (k K8sClient) WatchBlank(ctx context.Context, namespace string) (chan InformEvent, error) {
	return make(chan InformEvent), nil
}

func (k K8sClient) Watch(ctx context.Context, namespace string) (chan InformEvent, error) {

	lw := cache.NewListWatchFromClient(k.clientset.Apps().RESTClient(), "deployments", namespace, fields.Everything())

	ch := make(chan InformEvent)
	_, controller := cache.NewInformer(
		lw, &v1.Deployment{}, 10*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ch <- AddInformEvent{Obj: obj}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				ch <- UpdateInformEvent{OldObj: oldObj, NewObj: newObj}
			},
			DeleteFunc: func(last interface{}) {
				ch <- DeleteInformEvent{Last: last}
			},
		})

	go controller.Run(ctx.Done())

	return ch, nil
}
