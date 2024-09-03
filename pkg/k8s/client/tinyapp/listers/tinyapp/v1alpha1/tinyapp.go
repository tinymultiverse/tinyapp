// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// TinyAppLister helps list TinyApps.
// All objects returned here must be treated as read-only.
type TinyAppLister interface {
	// List lists all TinyApps in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.TinyApp, err error)
	// TinyApps returns an object that can list and get TinyApps.
	TinyApps(namespace string) TinyAppNamespaceLister
	TinyAppListerExpansion
}

// tinyAppLister implements the TinyAppLister interface.
type tinyAppLister struct {
	indexer cache.Indexer
}

// NewTinyAppLister returns a new TinyAppLister.
func NewTinyAppLister(indexer cache.Indexer) TinyAppLister {
	return &tinyAppLister{indexer: indexer}
}

// List lists all TinyApps in the indexer.
func (s *tinyAppLister) List(selector labels.Selector) (ret []*v1alpha1.TinyApp, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.TinyApp))
	})
	return ret, err
}

// TinyApps returns an object that can list and get TinyApps.
func (s *tinyAppLister) TinyApps(namespace string) TinyAppNamespaceLister {
	return tinyAppNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// TinyAppNamespaceLister helps list and get TinyApps.
// All objects returned here must be treated as read-only.
type TinyAppNamespaceLister interface {
	// List lists all TinyApps in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.TinyApp, err error)
	// Get retrieves the TinyApp from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.TinyApp, error)
	TinyAppNamespaceListerExpansion
}

// tinyAppNamespaceLister implements the TinyAppNamespaceLister
// interface.
type tinyAppNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all TinyApps in the indexer for a given namespace.
func (s tinyAppNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.TinyApp, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.TinyApp))
	})
	return ret, err
}

// Get retrieves the TinyApp from the indexer for a given namespace and name.
func (s tinyAppNamespaceLister) Get(name string) (*v1alpha1.TinyApp, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("tinyapp"), name)
	}
	return obj.(*v1alpha1.TinyApp), nil
}