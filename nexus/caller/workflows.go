package caller

import (
	"errors"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/temporalio/samples-go/nexus/service"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue    = "my-caller-workflow-task-queue"
	endpointName = "myendpoint"
)

func NullCallerWorkflow(ctx workflow.Context, nexusOp string) (string, error) {

	logger := workflow.GetLogger(ctx)

	logger.Info("NullCallerWorkflow: START")

	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, nexusOp, nexus.NoValue(nil), workflow.NexusOperationOptions{})

	var res nexus.NoValue
	if err := fut.Get(ctx, &res); err != nil {
		logger.Info("NullCallerWorkflow: GET ERROR", err)
		var nexusError *temporal.NexusOperationError
		if errors.As(err, &nexusError) {
			logger.Error(fmt.Sprintf("nexusError: %#v\n", nexusError))
			var appError *temporal.ApplicationError
			if errors.As(nexusError.Cause, &appError) {
				logger.Error(fmt.Sprintf("appError: %#v\n", appError))
			}
		} else {
			logger.Error(fmt.Sprintf("unknownError: %#v\n", err))
		}
		return "", err
	}
	logger.Info("NullCallerWorkflow: OK")
	return "ok", nil
}

func EchoCallerWorkflow(ctx workflow.Context, message string) (string, error) {
	logger := workflow.GetLogger(ctx)

	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.EchoOperationName, service.EchoInput{Message: message}, workflow.NexusOperationOptions{})

	var res service.EchoOutput
	if err := fut.Get(ctx, &res); err != nil {
		var nexusError *temporal.NexusOperationError
		if errors.As(err, &nexusError) {
			logger.Error(fmt.Sprintf("nexusError: %#v\n", nexusError))
			var appError *temporal.ApplicationError
			if errors.As(nexusError.Cause, &appError) {
				logger.Error(fmt.Sprintf("appError: %#v\n", appError))
			}
		}
		return "", err
	}
	return res.Message, nil
}

func HelloCallerWorkflow(ctx workflow.Context, name string, language service.Language) (string, error) {
	logger := workflow.GetLogger(ctx)
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.HelloOperationName, service.HelloInput{Name: name, Language: language}, workflow.NexusOperationOptions{})
	var res service.HelloOutput

	// Optionally wait for the operation to be started. NexusOperationExecution will contain the operation ID in
	// case this operation is asynchronous.
	var exec workflow.NexusOperationExecution
	if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
		logger.Error("GetNexusOperationExecution")
		logger.Error(fmt.Sprintf("IsCancelled: %v", temporal.IsCanceledError(err)))
		var nexusError *temporal.NexusOperationError
		if errors.As(err, &nexusError) {
			logger.Error(fmt.Sprintf("nexusError: %#v\n", nexusError))
			cause := nexusError.Cause
			switch cause.(type) {
			case *temporal.ApplicationError:
				logger.Error(fmt.Sprintf("appError: %#v\n", cause))
			case *temporal.CanceledError:
				logger.Error(fmt.Sprintf("canceledError: %#v\n", cause))
			default:
				logger.Error(fmt.Sprintf("unknownError: %#v\n", cause))
			}
		}
		return "", err
	}

	if err := fut.Get(ctx, &res); err != nil {
		logger.Error(fmt.Sprintf("Result error %#v\n", err))
		logger.Error(fmt.Sprintf("IsCancelled: %v", temporal.IsCanceledError(err)))
		var nexusError *temporal.NexusOperationError
		if errors.As(err, &nexusError) {
			logger.Error(fmt.Sprintf("nexusError: %#v\n", nexusError))
			cause := nexusError.Cause
			switch cause.(type) {
			case *temporal.ApplicationError:
				logger.Error(fmt.Sprintf("appError: %#v\n", cause))
			case *temporal.CanceledError:
				logger.Error(fmt.Sprintf("canceledError: %#v\n", cause))
				return "cancelled", nil
			default:
				logger.Error(fmt.Sprintf("unknownError: %#v\n", cause))
			}
			/*
				var appError *temporal.ApplicationError
				if errors.As(nexusError.Cause, &appError) {
					logger.Error(fmt.Sprintf("appError: %#v\n", appError))
				}
			*/
		}
		return "", err
	}

	return res.Message, nil
}

func HelloCallerWorkflow2(ctx workflow.Context, name string, language service.Language) (string, error) {
	logger := workflow.GetLogger(ctx)
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.HelloOperation2Name, service.HelloInput{Name: name, Language: language}, workflow.NexusOperationOptions{})
	var res service.HelloOutput

	if err := fut.Get(ctx, &res); err != nil {
		logger.Error(fmt.Sprintf("Result error %#v\n", err))
		logger.Error(fmt.Sprintf("IsCancelled: %v", temporal.IsCanceledError(err)))
		var nexusError *temporal.NexusOperationError
		if errors.As(err, &nexusError) {
			logger.Error(fmt.Sprintf("nexusError: %#v\n", nexusError))
			cause := nexusError.Cause
			switch cause.(type) {
			case *temporal.ApplicationError:
				logger.Error(fmt.Sprintf("appError: %#v\n", cause))
			case *temporal.CanceledError:
				logger.Error(fmt.Sprintf("canceledError: %#v\n", cause))
				return "cancelled", nil
			default:
				logger.Error(fmt.Sprintf("unknownError: %#v\n", cause))
			}
		}
		return "", err
	}

	return res.Message, nil
}
