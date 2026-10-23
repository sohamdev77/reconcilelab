// Package simulator implements a deliberately narrow, sequential ConfigMap API.
package simulator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrUnsupported distinguishes fidelity limits from injected Kubernetes failures.
var ErrUnsupported = errors.New("simulator operation unsupported")

type Operation struct {
	Verb   string                 `json:"verb"`
	Key    string                 `json:"key"`
	Result string                 `json:"result"`
	Object map[string]interface{} `json:"object,omitempty"`
}

// Client embeds a rejecting client for unsupported methods, never a real cluster client.
type Client struct {
	client.Client
	Live       map[string]*unstructured.Unstructured
	Cache      map[string]*unstructured.Unstructured
	Frozen     bool
	Revision   int
	Faults     map[string][]string
	Operations []Operation
}
type unsupported struct{}

func (unsupported) Scheme() *runtime.Scheme { return runtime.NewScheme() }
func (unsupported) RESTMapper() meta.RESTMapper {
	return meta.NewDefaultRESTMapper([]schema.GroupVersion{{Version: "v1"}})
}
func (unsupported) GroupVersionKindFor(o runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, fmt.Errorf("%w: GVK lookup unsupported; set v1/ConfigMap explicitly", ErrUnsupported)
}
func (unsupported) IsObjectNamespaced(runtime.Object) (bool, error) { return true, nil }
func (unsupported) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return fmt.Errorf("%w: simulator: List unsupported", ErrUnsupported)
}
func (unsupported) Patch(context.Context, client.Object, client.Patch, ...client.PatchOption) error {
	return fmt.Errorf("%w: simulator: Patch unsupported", ErrUnsupported)
}
func (unsupported) DeleteAllOf(context.Context, client.Object, ...client.DeleteAllOfOption) error {
	return fmt.Errorf("%w: simulator: DeleteAllOf unsupported", ErrUnsupported)
}
func (unsupported) Status() client.SubResourceWriter            { return rejectSub{} }
func (unsupported) SubResource(string) client.SubResourceClient { return rejectSub{} }

type rejectSub struct{}

func (rejectSub) Get(context.Context, client.Object, client.Object, ...client.SubResourceGetOption) error {
	return fmt.Errorf("%w: subresources unsupported", ErrUnsupported)
}
func (rejectSub) Create(context.Context, client.Object, client.Object, ...client.SubResourceCreateOption) error {
	return fmt.Errorf("%w: subresources unsupported", ErrUnsupported)
}
func (rejectSub) Update(context.Context, client.Object, ...client.SubResourceUpdateOption) error {
	return fmt.Errorf("%w: subresources unsupported", ErrUnsupported)
}
func (rejectSub) Patch(context.Context, client.Object, client.Patch, ...client.SubResourcePatchOption) error {
	return fmt.Errorf("%w: subresources unsupported", ErrUnsupported)
}
func New(objects []unstructured.Unstructured) *Client {
	c := &Client{Client: unsupported{}, Live: map[string]*unstructured.Unstructured{}, Faults: map[string][]string{}}
	for _, o := range objects {
		c.Revision++
		p := o.DeepCopy()
		p.SetResourceVersion(strconv.Itoa(c.Revision))
		c.Live[key(p)] = p
	}
	c.Sync()
	return c
}
func key(o client.Object) string { return o.GetNamespace() + "/" + o.GetName() }
func (c *Client) Sync() {
	c.Cache = map[string]*unstructured.Unstructured{}
	for k, o := range c.Live {
		c.Cache[k] = o.DeepCopy()
	}
	c.Frozen = false
}
func (c *Client) Snapshot() []unstructured.Unstructured {
	out := []unstructured.Unstructured{}
	keys := []string{}
	for k := range c.Live {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, *c.Live[k].DeepCopy())
	}
	return out
}
func (c *Client) fault(verb, k string) error {
	q := c.Faults[verb]
	if len(q) == 0 {
		return nil
	}
	c.Faults[verb] = q[1:]
	if q[0] == "conflict" {
		return apierrors.NewConflict(schema.GroupResource{Resource: "configmaps"}, k, fmt.Errorf("injected conflict"))
	}
	return apierrors.NewServiceUnavailable("injected unavailable")
}
func (c *Client) record(v, k string, o client.Object, e error) {
	r := "ok"
	if e != nil {
		r = e.Error()
	}
	op := Operation{Verb: v, Key: k, Result: r}
	if o != nil {
		b, _ := json.Marshal(o)
		_ = json.Unmarshal(b, &op.Object)
	}
	c.Operations = append(c.Operations, op)
}
func supported(o client.Object) error {
	if u, ok := o.(*unstructured.Unstructured); ok {
		if u.GetKind() != "ConfigMap" || u.GetAPIVersion() != "v1" {
			return fmt.Errorf("%w: only v1 ConfigMap supported", ErrUnsupported)
		}
	} else if _, ok := o.(*corev1.ConfigMap); !ok {
		return fmt.Errorf("%w: only ConfigMap supported", ErrUnsupported)
	}
	return nil
}
func (c *Client) Get(ctx context.Context, k client.ObjectKey, o client.Object, opts ...client.GetOption) (err error) {
	defer func() { c.record("get", k.String(), o, err) }()
	if err = supported(o); err != nil {
		return
	}
	if len(opts) > 0 {
		return fmt.Errorf("%w: Get options unsupported", ErrUnsupported)
	}
	if err = c.fault("get", k.String()); err != nil {
		return
	}
	p := c.Cache[k.String()]
	if p == nil {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, k.Name)
	}
	switch target := o.(type) {
	case *unstructured.Unstructured:
		target.Object = p.DeepCopy().Object
	case *corev1.ConfigMap:
		var decoded corev1.ConfigMap
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &decoded); err != nil {
			return err
		}
		*target = decoded
	}
	return nil
}
func (c *Client) write(v string, o client.Object) (err error) {
	defer func() { c.record(v, key(o), o, err) }()
	if err = supported(o); err != nil {
		return
	}
	if err = c.fault(v, key(o)); err != nil {
		return
	}
	old := c.Live[key(o)]
	switch v {
	case "create":
		if old != nil {
			return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, o.GetName())
		}
	case "update", "delete":
		if old == nil {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, o.GetName())
		}
		if v == "update" && o.GetResourceVersion() != old.GetResourceVersion() {
			return apierrors.NewConflict(schema.GroupResource{Resource: "configmaps"}, o.GetName(), fmt.Errorf("stale resourceVersion"))
		}
	}
	c.Revision++
	if v == "delete" {
		delete(c.Live, key(o))
	} else {
		b, e := json.Marshal(o)
		if e != nil {
			return e
		}
		p := &unstructured.Unstructured{}
		if e = json.Unmarshal(b, &p.Object); e != nil {
			return e
		}
		p.SetAPIVersion("v1")
		p.SetKind("ConfigMap")
		p.SetResourceVersion(strconv.Itoa(c.Revision))
		o.SetResourceVersion(p.GetResourceVersion())
		c.Live[key(o)] = p
	}
	if !c.Frozen {
		c.Sync()
	}
	return nil
}
func (c *Client) Create(ctx context.Context, o client.Object, opts ...client.CreateOption) error {
	if len(opts) > 0 {
		return fmt.Errorf("%w: Create options unsupported", ErrUnsupported)
	}
	return c.write("create", o)
}
func (c *Client) Update(ctx context.Context, o client.Object, opts ...client.UpdateOption) error {
	if len(opts) > 0 {
		return fmt.Errorf("%w: Update options unsupported", ErrUnsupported)
	}
	return c.write("update", o)
}
func (c *Client) Delete(ctx context.Context, o client.Object, opts ...client.DeleteOption) error {
	if len(opts) > 0 {
		return fmt.Errorf("%w: Delete options unsupported", ErrUnsupported)
	}
	return c.write("delete", o)
}

func (unsupported) Apply(context.Context, runtime.ApplyConfiguration, ...client.ApplyOption) error {
	return fmt.Errorf("%w: simulator: Apply unsupported", ErrUnsupported)
}
func (rejectSub) Apply(context.Context, runtime.ApplyConfiguration, ...client.SubResourceApplyOption) error {
	return fmt.Errorf("%w: subresources unsupported", ErrUnsupported)
}
func (unsupported) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return fmt.Errorf("%w: Get unsupported", ErrUnsupported)
}
func (unsupported) Create(context.Context, client.Object, ...client.CreateOption) error {
	return fmt.Errorf("%w: Create unsupported", ErrUnsupported)
}
func (unsupported) Update(context.Context, client.Object, ...client.UpdateOption) error {
	return fmt.Errorf("%w: Update unsupported", ErrUnsupported)
}
func (unsupported) Delete(context.Context, client.Object, ...client.DeleteOption) error {
	return fmt.Errorf("%w: Delete unsupported", ErrUnsupported)
}
