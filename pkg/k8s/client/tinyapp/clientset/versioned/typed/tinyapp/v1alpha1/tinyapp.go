// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	scheme "github.com/tinymultiverse/tinyapp/pkg/k8s/client/tinyapp/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TinyAppsGetter has a method to return a TinyAppInterface.
// A group's client should implement this interface.
type TinyAppsGetter interface {
	TinyApps(namespace string) TinyAppInterface
}

// TinyAppInterface has methods to work with TinyApp resources.
type TinyAppInterface interface {
	Create(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.CreateOptions) (*v1alpha1.TinyApp, error)
	Update(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.UpdateOptions) (*v1alpha1.TinyApp, error)
	UpdateStatus(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.UpdateOptions) (*v1alpha1.TinyApp, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.TinyApp, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TinyAppList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TinyApp, err error)
	TinyAppExpansion
}

// tinyApps implements TinyAppInterface
type tinyApps struct {
	client rest.Interface
	ns     string
}

// newTinyApps returns a TinyApps
func newTinyApps(c *TinymultiverseV1alpha1Client, namespace string) *tinyApps {
	return &tinyApps{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the tinyApp, and returns the corresponding tinyApp object, and an error if there is any.
func (c *tinyApps) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TinyApp, err error) {
	result = &v1alpha1.TinyApp{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("tinyapps").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TinyApps that match those selectors.
func (c *tinyApps) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TinyAppList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.TinyAppList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("tinyapps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested tinyApps.
func (c *tinyApps) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("tinyapps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a tinyApp and creates it.  Returns the server's representation of the tinyApp, and an error, if there is any.
func (c *tinyApps) Create(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.CreateOptions) (result *v1alpha1.TinyApp, err error) {
	result = &v1alpha1.TinyApp{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("tinyapps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tinyApp).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a tinyApp and updates it. Returns the server's representation of the tinyApp, and an error, if there is any.
func (c *tinyApps) Update(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.UpdateOptions) (result *v1alpha1.TinyApp, err error) {
	result = &v1alpha1.TinyApp{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("tinyapps").
		Name(tinyApp.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tinyApp).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *tinyApps) UpdateStatus(ctx context.Context, tinyApp *v1alpha1.TinyApp, opts v1.UpdateOptions) (result *v1alpha1.TinyApp, err error) {
	result = &v1alpha1.TinyApp{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("tinyapps").
		Name(tinyApp.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tinyApp).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the tinyApp and deletes it. Returns an error if one occurs.
func (c *tinyApps) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("tinyapps").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *tinyApps) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("tinyapps").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched tinyApp.
func (c *tinyApps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TinyApp, err error) {
	result = &v1alpha1.TinyApp{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("tinyapps").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
