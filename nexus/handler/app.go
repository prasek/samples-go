package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/service"
)

var NullSyncOp = temporalnexus.NewSyncOperation(
	service.NullSyncOp,
	func(ctx context.Context, c client.Client, input nexus.NoValue, options nexus.StartOperationOptions) (nexus.NoValue, error) {
		logger := temporalnexus.GetLogger(ctx)
		logger.Info("NullSyncOp start ...")
		return nil, nil
	})

var NullAsyncOp = temporalnexus.NewWorkflowRunOperation(
	service.NullAsyncOp,
	NullWorkflow,
	func(ctx context.Context, input nexus.NoValue, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		logger := temporalnexus.GetLogger(ctx)
		logger.Info("NullAsyncOp start ...")
		return client.StartWorkflowOptions{
			// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
			// For this example, we're using the request ID allocated by Temporal when the caller workflow schedules
			// the operation, this ID is guaranteed to be stable across retries of this operation.
			ID: options.RequestID,
			// Task queue defaults to the task queue this operation is handled on.
		}, nil
	})

func NullWorkflow(ctx workflow.Context, input nexus.NoValue) (nexus.NoValue, error) {
	ctx = workflow.WithActivityOptions(ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		},
	)

	var result string

	err := workflow.ExecuteActivity(ctx,
		HelloActivity,
		"test 123: activity",
	).Get(ctx, &result)

	if err != nil {
		return nil, err
	}

	return nil, nil

}

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = temporalnexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, c client.Client, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	time.Sleep(1 * time.Second)
	return service.EchoOutput(input), nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation2 = temporalnexus.NewWorkflowRunOperation(service.HelloOperation2Name, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	return client.StartWorkflowOptions{
		// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
		// For this example, we're using the request ID allocated by Temporal when the caller workflow schedules
		// the operation, this ID is guaranteed to be stable across retries of this operation.
		ID: options.RequestID,
		// Task queue defaults to the task queue this operation is handled on.
	}, nil
})

var HelloOperation, _ = temporalnexus.NewWorkflowRunOperationWithOptions(temporalnexus.WorkflowRunOperationOptions[service.HelloInput, *service.HelloOutput]{
	Name: service.HelloOperationName,
	Handler: func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (temporalnexus.WorkflowHandle[*service.HelloOutput], error) {
		logger := temporalnexus.GetLogger(ctx)
		logger.Info("HelloOperation start ...")

		wfOpts := client.StartWorkflowOptions{
			// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
			// For this example, we're using the request ID allocated by Temporal when the caller workflow schedules
			// the operation, this ID is guaranteed to be stable across retries of this operation.
			ID: options.RequestID,
			// Task queue defaults to the task queue this operation is handled on.
		}

		handle, err := temporalnexus.ExecuteWorkflow(ctx, options, wfOpts, HelloHandlerWorkflow, input)
		if err != nil {
			return nil, err
		}
		return handle, err
	}})

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	ctx = workflow.WithActivityOptions(ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		},
	)

	var result string

	err := workflow.ExecuteActivity(ctx,
		HelloActivity,
		"test 123: activity",
	).Get(ctx, &result)

	if err != nil {
		return service.HelloOutput{}, err
	}

	return service.HelloOutput{Message: result}, nil

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

func HelloActivity(ctx context.Context, msg string) (string, error) {
	return msg + " OK", nil
}
