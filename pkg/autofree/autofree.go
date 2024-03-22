package autofree

import (
	"context"
	"github.com/CloudNativeGame/aigc-gateway/pkg/options"
	"github.com/CloudNativeGame/aigc-gateway/pkg/resources"
	"github.com/CloudNativeGame/aigc-gateway/pkg/storage"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"
)

type AutoFree struct {
	opts *options.ServerOption
}

func NewAutoFree(serverOpts *options.ServerOption) *AutoFree {
	af := &AutoFree{
		opts: serverOpts,
	}
	return af
}

func (af *AutoFree) Run(ctx context.Context) {
	if af.opts.CheckInterval == 0 || af.opts.IdleLimit == 0 {
		return
	}
	go wait.Until(af.runOnce, af.opts.CheckInterval, ctx.Done())
}

func (af *AutoFree) runOnce() {
	rm := resources.NewResourceManager()
	resources, err := rm.ListResources(nil, nil)
	if err != nil {
		klog.Errorf("AutoFree ListResources failed: %v", err)
		return
	}
	for _, res := range resources {
		af.checkResource(res.GetNamespace(), res.GetName())
	}
}

func (af *AutoFree) checkResource(namespace, name string) {
	allStatus, err := storage.Get().GetAllStatus(context.Background(), namespace, name)
	if err != nil {
		return
	}
	if allStatus == nil {
		return
	}
	for id, status := range allStatus {
		if time.Since(status.Timestamp) < af.opts.IdleLimit {
			continue
		}
		// free it
		resourceManager := resources.NewResourceManager()
		rm := &resources.ResourceMeta{
			Namespace: namespace,
			Name:      name,
			ID:        id,
		}
		if err := resourceManager.PauseResource(rm); err != nil {
			klog.Errorf("checkResource PauseResource %s:%s failed: %v", name, id, err)
			continue
		}
		storage.Get().DeleteRecord(context.Background(), rm)
		klog.Infof("AutoFreeResource %s id %s", rm.Name, rm.ID)
	}
}

func Run(ctx context.Context, serverOpts *options.ServerOption) {
	af := NewAutoFree(serverOpts)
	af.Run(ctx)
}
