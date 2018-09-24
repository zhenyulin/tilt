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
	obj interface{}
}

type UpdateInformEvent struct {
	oldObj interface{}
	newObj interface{}
}

type DeleteInformEvent struct {
	last interface{}
}

func (AddInformEvent) informEvent()    {}
func (UpdateInformEvent) informEvent() {}
func (DeleteInformEvent) informEvent() {}

func (k K8sClient) Watch(ctx context.Context, namespace string) (chan InformEvent, error) {
	lw := cache.NewListWatchFromClient(k.core.RESTClient(), "deployments", namespace, fields.Everything())

	ch := make(chan InformEvent)
	_, controller := cache.NewInformer(
		lw, &v1.Deployment{}, 10*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ch <- AddInformEvent{obj: obj}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				ch <- UpdateInformEvent{oldObj: oldObj, newObj: newObj}
			},
			DeleteFunc: func(last interface{}) {
				ch <- DeleteInformEvent{last: last}
			},
		})

	go controller.Run(ctx.Done())

	return ch, nil
}
