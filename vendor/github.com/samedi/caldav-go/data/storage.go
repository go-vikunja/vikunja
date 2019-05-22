package data

import (
	"github.com/samedi/caldav-go/errs"
	"github.com/samedi/caldav-go/files"
	"io/ioutil"
	"log"
	"os"
)

// Storage is the inteface responsible for the CRUD operations on the CalDAV resources. It represents
// where the resources should be fetched from and the various operations which can be performed on it.
// This is the interface one should implement in case it needs a custom storage strategy, like fetching
// data from the cloud, local DB, etc. After that, the custom storage implementation can be setup to be used
// in the server by passing the object instance to `caldav.SetupStorage`.
type Storage interface {
	// GetResources gets a list of resources based on a given `rpath`. The
	// `rpath` is the path to the original resource that's being requested. The resultant list
	// will/must contain that original resource in it, apart from any additional resources. It also receives
	// `withChildren` flag to say if the result must also include all the original resource`s
	// children (if original is a collection resource). If `true`, the result will have the requested resource + children.
	// If `false`, it will have only the requested original resource (from the `rpath` path).
	// It returns errors if anything went wrong or if it could not find any resource on `rpath` path.
	GetResources(rpath string, withChildren bool) ([]Resource, error)
	// GetResourcesByList fetches a list of resources by path from the storage.
	// This method fetches all the `rpaths` and return an array of the reosurces found.
	// No error 404 will be returned if one of the resources cannot be found.
	// Errors are returned if any errors other than "not found" happens.
	GetResourcesByList(rpaths []string) ([]Resource, error)
	// GetResourcesByFilters returns the filtered children of a target collection resource.
	// The target collection resource is the one pointed by the `rpath` parameter. All of its children
	// will be checked against a set of `filters` and the matching ones are returned. The results
	// contains only the filtered children and does NOT include the target resource. If the target resource
	// is not a collection, an empty array is returned as the result.
	GetResourcesByFilters(rpath string, filters *ResourceFilter) ([]Resource, error)
	// GetResource gets the requested resource based on a given `rpath` path. It returns the resource (if found) or
	// nil (if not found). Also returns a flag specifying if the resource was found or not.
	GetResource(rpath string) (*Resource, bool, error)
	// GetShallowResource has the same behaviour of `storage.GetResource`. The only difference is that, for collection resources,
	// it does not return its children in the collection `storage.Resource` struct (hence the name shallow). The motive is
	// for optimizations reasons, as this function is used on places where the collection's children are not important.
	GetShallowResource(rpath string) (*Resource, bool, error)
	// CreateResource creates a new resource on the `rpath` path with a given `content`.
	CreateResource(rpath, content string) (*Resource, error)
	// UpdateResource udpates a resource on the `rpath` path with a given `content`.
	UpdateResource(rpath, content string) (*Resource, error)
	// DeleteResource deletes a resource on the `rpath` path.
	DeleteResource(rpath string) error
}

// FileStorage is the storage that deals with resources as files in the file system. So, a collection resource
// is treated as a folder/directory and its children resources are the files it contains. Non-collection resources are just plain files.
// Each file represents then a CalAV resource and the data expects to contain the iCal data to feed the calendar events.
type FileStorage struct {
}

// GetResources get the file resources based on the `rpath`. See `Storage.GetResources` doc.
func (fs *FileStorage) GetResources(rpath string, withChildren bool) ([]Resource, error) {
	result := []Resource{}

	// tries to open the file by the given path
	f, e := fs.openResourceFile(rpath, os.O_RDONLY)
	if e != nil {
		return nil, e
	}

	// add it as a resource to the result list
	finfo, _ := f.Stat()
	resource := NewResource(rpath, &FileResourceAdapter{finfo, rpath})
	result = append(result, resource)

	// if the file is a dir, add its children to the result list
	if withChildren && finfo.IsDir() {
		dirFiles, _ := f.Readdir(0)
		for _, finfo := range dirFiles {
			childPath := files.JoinPaths(rpath, finfo.Name())
			resource = NewResource(childPath, &FileResourceAdapter{finfo, childPath})
			result = append(result, resource)
		}
	}

	return result, nil
}

// GetResourcesByFilters get the file resources based on the `rpath` and a set of filters. See `Storage.GetResourcesByFilters` doc.
func (fs *FileStorage) GetResourcesByFilters(rpath string, filters *ResourceFilter) ([]Resource, error) {
	result := []Resource{}

	childPaths := fs.getDirectoryChildPaths(rpath)
	for _, path := range childPaths {
		resource, _, err := fs.GetShallowResource(path)

		if err != nil {
			// if we can't find this resource, something weird went wrong, but not that serious, so we log it and continue
			log.Printf("WARNING: returned error when trying to get resource with path %s from collection with path %s. Error: %s", path, rpath, err)
			continue
		}

		// only add it if the resource matches the filters
		if filters == nil || filters.Match(resource) {
			result = append(result, *resource)
		}
	}

	return result, nil
}

// GetResourcesByList get a list of file resources based on a list of `rpaths`. See `Storage.GetResourcesByList` doc.
func (fs *FileStorage) GetResourcesByList(rpaths []string) ([]Resource, error) {
	results := []Resource{}

	for _, rpath := range rpaths {
		resource, found, err := fs.GetShallowResource(rpath)

		if err != nil && err != errs.ResourceNotFoundError {
			return nil, err
		}

		if found {
			results = append(results, *resource)
		}
	}

	return results, nil
}

// GetResource fetches and returns a single resource for a `rpath`. See `Storage.GetResoure` doc.
func (fs *FileStorage) GetResource(rpath string) (*Resource, bool, error) {
	// For simplicity we just return the shallow resource.
	return fs.GetShallowResource(rpath)
}

// GetShallowResource fetches and returns a single resource file/directory without any related children. See `Storage.GetShallowResource` doc.
func (fs *FileStorage) GetShallowResource(rpath string) (*Resource, bool, error) {
	resources, err := fs.GetResources(rpath, false)

	if err != nil {
		return nil, false, err
	}

	if resources == nil || len(resources) == 0 {
		return nil, false, errs.ResourceNotFoundError
	}

	res := resources[0]
	return &res, true, nil
}

// CreateResource creates a file resource with the provided `content`. See `Storage.CreateResource` doc.
func (fs *FileStorage) CreateResource(rpath, content string) (*Resource, error) {
	rAbsPath := files.AbsPath(rpath)

	if fs.isResourcePresent(rAbsPath) {
		return nil, errs.ResourceAlreadyExistsError
	}

	// create parent directories (if needed)
	if err := os.MkdirAll(files.DirPath(rAbsPath), os.ModePerm); err != nil {
		return nil, err
	}

	// create file/resource and write content
	f, err := os.Create(rAbsPath)
	if err != nil {
		return nil, err
	}
	f.WriteString(content)

	finfo, _ := f.Stat()
	res := NewResource(rpath, &FileResourceAdapter{finfo, rpath})
	return &res, nil
}

// UpdateResource updates a file resource with the provided `content`. See `Storage.UpdateResource` doc.
func (fs *FileStorage) UpdateResource(rpath, content string) (*Resource, error) {
	f, e := fs.openResourceFile(rpath, os.O_RDWR)
	if e != nil {
		return nil, e
	}

	// update content
	f.Truncate(0)
	f.WriteString(content)

	finfo, _ := f.Stat()
	res := NewResource(rpath, &FileResourceAdapter{finfo, rpath})
	return &res, nil
}

// DeleteResource deletes a file resource (and possibly all its children in case of a collection). See `Storage.DeleteResource` doc.
func (fs *FileStorage) DeleteResource(rpath string) error {
	err := os.Remove(files.AbsPath(rpath))

	return err
}

func (fs *FileStorage) isResourcePresent(rpath string) bool {
	_, found, _ := fs.GetShallowResource(rpath)

	return found
}

func (fs *FileStorage) openResourceFile(filepath string, mode int) (*os.File, error) {
	f, e := os.OpenFile(files.AbsPath(filepath), mode, 0666)
	if e != nil {
		if os.IsNotExist(e) {
			return nil, errs.ResourceNotFoundError
		}
		return nil, e
	}

	return f, nil
}

func (fs *FileStorage) getDirectoryChildPaths(dirpath string) []string {
	content, err := ioutil.ReadDir(files.AbsPath(dirpath))
	if err != nil {
		log.Printf("ERROR: Could not read resource as file directory.\nError: %s.\nResource path: %s.", err, dirpath)
		return nil
	}

	result := []string{}
	for _, file := range content {
		fpath := files.JoinPaths(dirpath, file.Name())
		result = append(result, fpath)
	}

	return result
}
