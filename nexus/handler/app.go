// @@@SNIPSTART samples-go-nexus-handler
package handler

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/ctxpropagation"
	"github.com/temporalio/samples-go/nexus/service"
)

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = temporalnexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, c client.Client, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// The method is provided with an SDK client that can be used for arbitrary calls such as signaling, querying,
	// and listing workflows but implementations are free to make arbitrary calls to other services or databases, or
	// perform simple computations such as this one.
	return service.EchoOutput(input), nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
/*
var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	fmt.Printf("header[test]=%s\n", options.Header.Get("test"))
	return client.StartWorkflowOptions{
		// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
		// For this example, we're using the request ID allocated by Temporal when the caller workflow schedules
		// the operation, this ID is guaranteed to be stable across retries of this operation.
		ID: options.RequestID,
		// Task queue defaults to the task queue this operation is handled on.
	}, nil
})
*/

var HelloOperation, _ = temporalnexus.NewWorkflowRunOperationWithOptions(temporalnexus.WorkflowRunOperationOptions[service.HelloInput, *service.HelloOutput]{
	Name: service.HelloOperationName,
	Handler: func(ctx context.Context, hi service.HelloInput, soo nexus.StartOperationOptions) (temporalnexus.WorkflowHandle[*service.HelloOutput], error) {

		temporalnexus.GetLogger(ctx).Info("custom header propagated from caller to Nexus handler", "nexus-test-header", soo.Header.Get("nexus-test-header"))

		ctx = context.WithValue(ctx, ctxpropagation.PropagateKey, &ctxpropagation.Values{Key: "nexus-test-header", Value: soo.Header.Get("nexus-test-header")})

		wfOpts := client.StartWorkflowOptions{
			ID: soo.RequestID,
		}

		handle, err := temporalnexus.ExecuteWorkflow(ctx, soo, wfOpts, HelloHandlerWorkflow, hi)
		if err != nil {
			return nil, err
		}
		return handle, nil
	},
})

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	if val := ctx.Value(ctxpropagation.PropagateKey); val != nil {
		vals := val.(ctxpropagation.Values)
		workflow.GetLogger(ctx).Info("custom context propagated to workflow", vals.Key, vals.Value)
	}
	switch input.Language {
	case service.EN:
		return service.HelloOutput{Message: "Hello " + input.Name + " 👋"}, nil
	case service.FR:
		return service.HelloOutput{Message: "Bonjour " + input.Name + " 👋"}, nil
	case service.DE:
		return service.HelloOutput{Message: "Hallo " + input.Name + " 👋"}, nil
	case service.ES:
		return service.HelloOutput{Message: "¡Hola! " + input.Name + " 👋"}, nil
	case service.TR:
		return service.HelloOutput{Message: "Merhaba " + input.Name + " 👋"}, nil
	}
	return service.HelloOutput{}, fmt.Errorf("unsupported language %q", input.Language)
}

// @@@SNIPEND
