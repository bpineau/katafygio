package main

import (
	"fmt"
	"os"

	"github.com/bpineau/katafygio/cmd"

	// import controllers so their init() is called
	_ "github.com/bpineau/katafygio/pkg/controllers/clusterrolebinding"
	_ "github.com/bpineau/katafygio/pkg/controllers/configmap"
	_ "github.com/bpineau/katafygio/pkg/controllers/cronjob"
	_ "github.com/bpineau/katafygio/pkg/controllers/daemonset"
	_ "github.com/bpineau/katafygio/pkg/controllers/deployment"
	_ "github.com/bpineau/katafygio/pkg/controllers/horizontalpodautoscaler"
	_ "github.com/bpineau/katafygio/pkg/controllers/ingress"
	_ "github.com/bpineau/katafygio/pkg/controllers/job"
	_ "github.com/bpineau/katafygio/pkg/controllers/limitrange"
	_ "github.com/bpineau/katafygio/pkg/controllers/namespace"
	_ "github.com/bpineau/katafygio/pkg/controllers/networkpolicy"
	_ "github.com/bpineau/katafygio/pkg/controllers/persistentvolume"
	_ "github.com/bpineau/katafygio/pkg/controllers/persistentvolumeclaim"
	_ "github.com/bpineau/katafygio/pkg/controllers/pod"
	_ "github.com/bpineau/katafygio/pkg/controllers/poddisruptionbudget"
	_ "github.com/bpineau/katafygio/pkg/controllers/podsecuritypolicy"
	_ "github.com/bpineau/katafygio/pkg/controllers/podtemplate"
	_ "github.com/bpineau/katafygio/pkg/controllers/replicationcontroller"
	_ "github.com/bpineau/katafygio/pkg/controllers/resourcequota"
	_ "github.com/bpineau/katafygio/pkg/controllers/rolebinding"
	_ "github.com/bpineau/katafygio/pkg/controllers/secret"
	_ "github.com/bpineau/katafygio/pkg/controllers/service"
	_ "github.com/bpineau/katafygio/pkg/controllers/serviceaccount"
	_ "github.com/bpineau/katafygio/pkg/controllers/statefulset"
	_ "github.com/bpineau/katafygio/pkg/controllers/storageclass"
	// ReplicaSets may consume a lot of memory, and is already subsumed
	// by Deployments (does anybody creates naked ReplicaSets by hand ?).
	//_ "github.com/bpineau/katafygio/pkg/controllers/replicaset"
)

//var privateExitHandler func(code int) = os.Exit
var privateExitHandler = os.Exit

// ExitWrapper allow unit tests on main() exit values
func ExitWrapper(exit int) {
	privateExitHandler(exit)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%+v", err)
		ExitWrapper(1)
	}
}
