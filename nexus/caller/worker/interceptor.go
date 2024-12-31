package main

import (
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

type nexusInterceptor struct {
	interceptor.WorkerInterceptorBase
	interceptor.WorkflowInboundInterceptorBase
	interceptor.WorkflowOutboundInterceptorBase
}

func (i *nexusInterceptor) InterceptWorkflow(
	ctx workflow.Context,
	next interceptor.WorkflowInboundInterceptor,
) interceptor.WorkflowInboundInterceptor {
	i.WorkflowInboundInterceptorBase.Next = next
	return i
}

func (i *nexusInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	i.WorkflowOutboundInterceptorBase.Next = outbound
	return i.WorkflowInboundInterceptorBase.Next.Init(i)
}

func (i *nexusInterceptor) ExecuteNexusOperation(
	ctx workflow.Context,
	input interceptor.ExecuteNexusOperationInput,
) workflow.NexusOperationFuture {
	input.NexusHeader["nexus-test-header"] = "present"
	return i.WorkflowOutboundInterceptorBase.Next.ExecuteNexusOperation(ctx, input)
}
