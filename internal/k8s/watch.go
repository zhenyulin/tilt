package k8s

type informEvent interface {
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

func (k k8sClient) Watch(ctx context.Context, namespace string) (chan interface{}, error) {
	lw := cache.NewListWatchFromClient(k.core.RESTClient(), v1.PodResource, namespace, fields.Everything())

	ch := make(chan informEvent)
	_, controller := cache.NewInformer(
		lw, &v1.Deployment{}, 10*time.Second,
		ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ch <- AddInformEvent{obj: obj}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				ch <- UpdateInformEvent{oldObj: oldObj, newObj: newObj}
			},
			DeleteFunc: func(obj interface{}) {
				ch <- DeleteInformEvent{obj: obj}
			},
		})

	go controller.Run(ctx.Done())

	return ch
}
