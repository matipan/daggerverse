// Code generated by dagger. DO NOT EDIT.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"main/internal/dagger"
	"main/internal/querybuilder"
	"main/internal/telemetry"
)

var dag = dagger.Connect()

func Tracer() trace.Tracer {
	return otel.Tracer("dagger.io/sdk.go")
}

// used for local MarshalJSON implementations
var marshalCtx = context.Background()

// called by main()
func setMarshalContext(ctx context.Context) {
	marshalCtx = ctx
	dagger.SetMarshalContext(ctx)
}

type DaggerObject = querybuilder.GraphQLMarshaller

type ExecError = dagger.ExecError

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// convertSlice converts a slice of one type to a slice of another type using a
// converter function
func convertSlice[I any, O any](in []I, f func(I) O) []O {
	out := make([]O, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func (r ImageUpdater) MarshalJSON() ([]byte, error) {
	var concrete struct{}
	return json.Marshal(&concrete)
}

func (r *ImageUpdater) UnmarshalJSON(bs []byte) error {
	var concrete struct{}
	err := json.Unmarshal(bs, &concrete)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()

	// Direct slog to the new stderr. This is only for dev time debugging, and
	// runtime errors/warnings.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	if err := dispatch(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func dispatch(ctx context.Context) error {
	ctx = telemetry.InitEmbedded(ctx, resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("dagger-go-sdk"),
		// TODO version?
	))
	defer telemetry.Close()

	// A lot of the "work" actually happens when we're marshalling the return
	// value, which entails getting object IDs, which happens in MarshalJSON,
	// which has no ctx argument, so we use this lovely global variable.
	setMarshalContext(ctx)

	fnCall := dag.CurrentFunctionCall()
	parentName, err := fnCall.ParentName(ctx)
	if err != nil {
		return fmt.Errorf("get parent name: %w", err)
	}
	fnName, err := fnCall.Name(ctx)
	if err != nil {
		return fmt.Errorf("get fn name: %w", err)
	}
	parentJson, err := fnCall.Parent(ctx)
	if err != nil {
		return fmt.Errorf("get fn parent: %w", err)
	}
	fnArgs, err := fnCall.InputArgs(ctx)
	if err != nil {
		return fmt.Errorf("get fn args: %w", err)
	}

	inputArgs := map[string][]byte{}
	for _, fnArg := range fnArgs {
		argName, err := fnArg.Name(ctx)
		if err != nil {
			return fmt.Errorf("get fn arg name: %w", err)
		}
		argValue, err := fnArg.Value(ctx)
		if err != nil {
			return fmt.Errorf("get fn arg value: %w", err)
		}
		inputArgs[argName] = []byte(argValue)
	}

	result, err := invoke(ctx, []byte(parentJson), parentName, fnName, inputArgs)
	if err != nil {
		return fmt.Errorf("invoke: %w", err)
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err = fnCall.ReturnValue(ctx, dagger.JSON(resultBytes)); err != nil {
		return fmt.Errorf("store return value: %w", err)
	}
	return nil
}
func invoke(ctx context.Context, parentJSON []byte, parentName string, fnName string, inputArgs map[string][]byte) (_ any, err error) {
	_ = inputArgs
	switch parentName {
	case "ImageUpdater":
		switch fnName {
		case "Update":
			var parent ImageUpdater
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var repo string
			if inputArgs["repo"] != nil {
				err = json.Unmarshal([]byte(inputArgs["repo"]), &repo)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg repo", err))
				}
			}
			var branch string
			if inputArgs["branch"] != nil {
				err = json.Unmarshal([]byte(inputArgs["branch"]), &branch)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg branch", err))
				}
			}
			var deployFilepath string
			if inputArgs["deployFilepath"] != nil {
				err = json.Unmarshal([]byte(inputArgs["deployFilepath"]), &deployFilepath)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg deployFilepath", err))
				}
			}
			var imageUrl string
			if inputArgs["imageUrl"] != nil {
				err = json.Unmarshal([]byte(inputArgs["imageUrl"]), &imageUrl)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg imageUrl", err))
				}
			}
			var gitUser string
			if inputArgs["gitUser"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gitUser"]), &gitUser)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gitUser", err))
				}
			}
			var gitEmail string
			if inputArgs["gitEmail"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gitEmail"]), &gitEmail)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gitEmail", err))
				}
			}
			var gitPassword *any
			if inputArgs["gitPassword"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gitPassword"]), &gitPassword)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gitPassword", err))
				}
			}
			var forceWithLease bool
			if inputArgs["forceWithLease"] != nil {
				err = json.Unmarshal([]byte(inputArgs["forceWithLease"]), &forceWithLease)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg forceWithLease", err))
				}
			}
			return nil, (*ImageUpdater).Update(&parent, ctx, repo, branch, deployFilepath, imageUrl, gitUser, gitEmail, gitPassword, forceWithLease)
		default:
			return nil, fmt.Errorf("unknown function %s", fnName)
		}
	case "":
		return dag.Module().
			WithObject(
				dag.TypeDef().WithObject("ImageUpdater", dagger.TypeDefWithObjectOpts{Description: "image-updater is a dagger module that updates, commits and pushes a kubernetes deployment file with a new image-url.\n\nThere are many alternatives to doing GitOps with kubernetes now a days, to name a few:\n\n- Automatically update the deployment with the new image using Kubectl or hitting the kubernetes API directly (no specific trace of the image deployed in the git repository)\n- Use tools such as Flux or ArgoCD to automatically watch a registry and deploy new images when they appear\n- Use Flux or ArgoCD as well but instead have them look for changes on specific manifests in a repository\n\nThis module is useful for the last alternative. When you have CD tools that are watching kubernetes manifests on your\nrepository you would need to change them explicitly. If you use Github or Gitlab there are actions that you can use to make\nthis changes (for example, there is a yq action and a git-auto-commit action), but the problem is that those workflows\ncannot be tested locally and they become complicated. In the case of github actions, if you run your action as part of a\nworkflow that takes a long time to run, it might happen that a new commit showed up and your push will fail. Solving this\nis possible, but it requires adding even more untestable bash. This is why this module exists. With image-updater you can\nimplement this logic in a single step that is reproducible locally."}).
					WithFunction(
						dag.Function("Update",
							dag.TypeDef().WithKind(dagger.VoidKind).WithOptional(true)).
							WithDescription("Update updates the kubernetes deployment file in the specified repository\nwith the new image URL.\nNOTE: this pushes a commit to your repository so make sure that you either\ndon't have a cyclic workflow trigger or that you use a token that prevents\nthis from happening.\n+optional forceWithLease").
							WithArg("repo", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("branch", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("deployFilepath", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("imageUrl", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gitUser", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gitEmail", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gitPassword", dag.TypeDef().WithKind(dagger.VoidKind).WithOptional(true)).
							WithArg("forceWithLease", dag.TypeDef().WithKind(dagger.BooleanKind)))), nil
	default:
		return nil, fmt.Errorf("unknown object %s", parentName)
	}
}
