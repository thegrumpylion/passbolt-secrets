package main

import (
	"io/ioutil"
	"time"

	"github.com/alexflint/go-arg"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	psclientset "github.com/thegrumpylion/passbolt-secrets/pkg/client/clientset/versioned"
	psinformers "github.com/thegrumpylion/passbolt-secrets/pkg/client/informers/externalversions"
	"github.com/thegrumpylion/passbolt-secrets/pkg/controller"
	"github.com/thegrumpylion/passbolt-secrets/pkg/signals"
)

func main() {
	args := &struct {
		MasterURL           string
		KubeConfig          string
		PassboltAddress     string
		PassboltFingerprint string
		PassboltPassword    string
		PassboltKeyFile     string
	}{}

	arg.MustParse(args)

	klog.InitFlags(nil)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(args.MasterURL, args.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	psClient, err := psclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building ps clientset: %s", err.Error())
	}

	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, time.Second*30)
	psInformerFactory := psinformers.NewSharedInformerFactory(psClient, time.Second*30)

	keyData, err := ioutil.ReadFile(args.PassboltKeyFile)
	if err != nil {
		klog.Fatalf("Error reading passbolt key file: %s", err.Error())
	}

	controller, err := controller.New(controller.KubeConfig{
		KubeClientset:          kubeClient,
		PassboltClientset:      psClient,
		SecretInformer:         kubeInformerFactory.Core().V1().Secrets(),
		PassboltSecretInformer: psInformerFactory.Passboltsecrets().V1alpha1().PassboltSecrets(),
	}, controller.PassboltConfig{
		Address:           args.PassboltAddress,
		ServerFingerprint: args.PassboltFingerprint,
		UserPassword:      args.PassboltPassword,
		UserPrivKey:       string(keyData),
	})
	if err != nil {
		klog.Fatalf("Error creating new controller: %s", err.Error())
	}

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	psInformerFactory.Start(stopCh)

	if err = controller.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}

}
