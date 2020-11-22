package controller

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"git.lastassault.de/sup/passbolt"

	psv1alpha1 "github.com/thegrumpylion/passbolt-secrets/pkg/apis/passboltsecrets/v1alpha1"
	clientset "github.com/thegrumpylion/passbolt-secrets/pkg/client/clientset/versioned"
	"github.com/thegrumpylion/passbolt-secrets/pkg/client/clientset/versioned/scheme"
	psscheme "github.com/thegrumpylion/passbolt-secrets/pkg/client/clientset/versioned/scheme"
	informers "github.com/thegrumpylion/passbolt-secrets/pkg/client/informers/externalversions/passboltsecrets/v1alpha1"
	listers "github.com/thegrumpylion/passbolt-secrets/pkg/client/listers/passboltsecrets/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const controllerAgentName = "passbolt-secrets-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a PassboltSecret is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a PassboltSecret fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by PassboltSecret"
	// MessageResourceSynced is the message used for an Event fired when a PassboltSecret
	// is synced successfully
	MessageResourceSynced = "PassboltSecret synced successfully"
)

type KubeConfig struct {
	KubeClientset          kubernetes.Interface
	PassboltClientset      clientset.Interface
	SecretInformer         coreinformers.SecretInformer
	PassboltSecretInformer informers.PassboltSecretInformer
}

type PassboltConfig struct {
	Address           string
	ServerFingerprint string
	UserPassword      string
	UserPrivKey       string
}

type Controller struct {
	kubeClientset kubernetes.Interface
	secretsLister corelisters.SecretLister
	secretsSynced cache.InformerSynced
	psClientset   clientset.Interface
	psLister      listers.PassboltSecretLister
	psSynced      cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
	recorder      record.EventRecorder

	client     *passbolt.Client
	privateKey []byte
	password   []byte
}

func New(kc KubeConfig, pc PassboltConfig) (*Controller, error) {

	utilruntime.Must(psscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kc.KubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientset: kc.KubeClientset,
		secretsLister: kc.SecretInformer.Lister(),
		secretsSynced: kc.SecretInformer.Informer().HasSynced,
		psClientset:   kc.PassboltClientset,
		psLister:      kc.PassboltSecretInformer.Lister(),
		psSynced:      kc.PassboltSecretInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PassboltSecrets"),
		recorder:      recorder,
	}

	if err := controller.passboltLogin(pc); err != nil {
		return nil, err
	}

	controller.privateKey = []byte(pc.UserPrivKey)
	controller.password = []byte(pc.UserPassword)

	klog.Info("Setting up event handlers")
	// Set up an event handler for when PassboltSecret resources change
	kc.PassboltSecretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePassboltSecret,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueuePassboltSecret(new)
		},
		DeleteFunc: controller.deletePassboltSecret,
	})

	return controller, nil
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting PassboltSecrets controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.secretsSynced, c.psSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {

	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the PassboltSecret resource with this namespace/name
	ps, err := c.psLister.PassboltSecrets(namespace).Get(name)
	if err != nil {
		// The PassboltSecret resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("passbolt secret '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	secName := ps.Spec.Name
	if secName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: secret name must be specified", key))
		return nil
	}

	secret, err := c.secretsLister.Secrets(ps.Namespace).Get(secName)
	if errors.IsNotFound(err) {
		sec, err := c.newSecret(ps)
		if err != nil {
			return err
		}
		secret, err = c.kubeClientset.CoreV1().Secrets(ps.Namespace).Create(context.TODO(), sec, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(secret, ps) {
		msg := fmt.Sprintf(MessageResourceExists, secret.Name)
		c.recorder.Event(ps, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = c.updatePassboltSecretStatus(ps, secret)
	if err != nil {
		return err
	}

	c.recorder.Event(ps, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil

}

func (c *Controller) updatePassboltSecretStatus(ps *psv1alpha1.PassboltSecret, secret *corev1.Secret) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	psCopy := ps.DeepCopy()
	psCopy.Status.Created = true
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.psClientset.PassboltsecretsV1alpha1().PassboltSecrets(ps.Namespace).Update(context.TODO(), psCopy, metav1.UpdateOptions{})
	return err
}

func (c *Controller) enqueuePassboltSecret(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) deletePassboltSecret(obj interface{}) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("object has no meta: %v", err))
		return
	}

	ps, err := c.psLister.PassboltSecrets(meta.GetNamespace()).Get(meta.GetName())
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	err = c.kubeClientset.CoreV1().Secrets(meta.GetNamespace()).Delete(context.TODO(), ps.Name, metav1.DeleteOptions{})
	if err != nil {
		utilruntime.HandleError(err)
	}
}

func (c *Controller) newSecret(ps *psv1alpha1.PassboltSecret) (*corev1.Secret, error) {

	var res *passbolt.Resource
	var sec *passbolt.Secret
	var err error

	if ps.Spec.Source.ID != "" {
		res, err = c.client.GetResource(context.TODO(), ps.Spec.Source.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get passbolt resource: %v", err)
		}
		sec, err = c.client.GetSecret(context.TODO(), ps.Spec.Source.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get passbolt secret: %v", err)
		}
	} else if ps.Spec.Source.Name != "" {
		resrs, err := c.client.GetResources(context.Background(), &passbolt.GetResourcesOptions{ContainSecret: true})
		if err != nil {
			return nil, fmt.Errorf("could not get passbolt resources: %v", err)
		}
		for _, r := range resrs {
			if r.Name == ps.Spec.Source.Name {
				res = &passbolt.Resource{}
				*res = r
				sec = &passbolt.Secret{}
				*sec = r.Secrets[0]
			}
		}
	}

	if res == nil || sec == nil {
		return nil, stderrors.New("resource or secret cannot be nil")
	}

	s, err := passbolt.DecryptMessage(c.privateKey, c.password, []byte(sec.Data))
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{}

	// set secret key and value
	secKey := "secret"
	if ps.Spec.SecretKey != "" {
		secKey = ps.Spec.SecretKey
	}
	data[secKey] = s

	// set username key and value if set
	if ps.Spec.UsernameKey != "" {
		if res.Username == "" {
			return nil, stderrors.New("Username expected to to have value but is empty")
		}
		data[ps.Spec.UsernameKey] = []byte(res.Username)
	}

	// set url key and value if set
	if ps.Spec.URLKey != "" {
		if res.URI == "" {
			return nil, stderrors.New("URL expected to to have value but is empty")
		}
		data[ps.Spec.URLKey] = []byte(res.URI)
	}

	// resource name
	name := res.Name
	if ps.Spec.Name != "" {
		name = ps.Spec.Name
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ps.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ps, psv1alpha1.SchemeGroupVersion.WithKind("PassboltSecret")),
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}, nil
}

func (c *Controller) passboltLogin(conf PassboltConfig) error {
	u, err := url.Parse(conf.Address)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := passbolt.NewClient(u, "", &http.Client{Transport: customTransport})

	passboltPublicKey, fingerprint, err := client.GetPublicKey(ctx)
	if err != nil {
		return err

	}

	if conf.ServerFingerprint != fingerprint {
		return stderrors.New("Passbolt Server FingerPrint is Wrong: " + fingerprint)
	}

	err = client.Login(ctx, []byte(conf.UserPrivKey), []byte(conf.UserPassword), []byte(passboltPublicKey))
	if err != nil {
		return err

	}

	c.client = client

	return nil
}
