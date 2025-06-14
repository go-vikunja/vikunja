# Models and services

The architecture of this web app is in general divided in two parts:
Models and services.

Services handle all "raw" requests, models contain data and methods to work with it.

A service takes (in most cases) a model and returns one.

## Table of Contents

* [Services](#services)
  * [Requests](#requests)
    * [Loading](#loading)
  * [Factories](#factories)
  * [Before Request](#before-request)
  * [After Request?](#after-request-)
* [Models](#models)
  * [Default Values](#default-values)
  * [Constructor](#constructor)
  * [Access Model Data](#access-to-the-data)

## Services

Services are located in `src/services`.

All services must inherit `AbstractService` which holds most of the methods.

A basic service can look like this:

```javascript
import AbstractService from './abstractService'
import ProjectModel from '../models/project'

export default class ProjectService extends AbstractService {
	constructor() {
		super({
			getAll: '/projects',
			get: '/projects/{id}',
			create: '/namespaces/{namespaceID}/projects',
			update: '/projects/{id}',
			delete: '/projects/{id}',
		})
	}
	
	modelFactory(data) {
		return new ProjectModel(data)
	}
}
```

The `constructor` calls its parent constructor and provides the paths to make the requests. 
The parent constructor will take these and save them in the service.
All paths are optional. Calling a method which doesn't have a path defined will fail. 

The placeholder values in the urls are replaced with the contents of variables with the same name in the corresponding model (the one you pass to the functions).

#### Requests

Several request types are possible:

| Name | HTTP Method |
|------|-------------|
| `get` | `GET` |
| `getAll` | `GET` |
| `create` | `PUT` |
| `update` | `POST` |
| `delete` | `DELETE` | 

Each method can take a model and optional url parameters as function parameters.
With the exception of `getAll()`, a model is always mandatory while parameters are not.

Each method returns a promise, so you can access a request result like so:

```javascript
service.getAll().then(result => {
	// Do something with result
})
```

The result is a ready-to-use model returned by the model factory.

##### Loading

Each service has a `loading` property, provided by `AbstractModel`.
This property is a `boolean`, it is automatically set to `true` (with a 100ms delay to avoid flickering) 
once the request is started and set to `false` once the request is finished.
You can use this to show and hide a loading animation in the frontend.

#### Factories

The `modelFactory` takes data, and returns a model. The result of all requests (with the exception
of the `delete` method) is run through this factory. The factory should return the appropriate model, see
[models](#models) down below on how to handle data in models.

`getAll()` checks if the response is an array, if that's the case, it will run each entry in it through
the `modelFactory`. 

It is possible to define a different factory for each request. This is done by implementing a method called 
`model{TYPE}Factory(data)` in your service. As a fallback if the specific factory is not defined, 
`modelFactory` will be used.

#### Before Request

For each request exists a `before{TYPE}(model)` method. It receives the model, can alter it and should return
the modified version.

This is useful to make unix timestamps from javascript dates, for example.

#### After Request ?

There is no `after{TYPE}` method which would be called after a request is done. 
Processing raw api data should be done in the constructor of the model, see more on that below.

## Models

Models are a bit simpler than services.
They usually consist of a declaration of defaults and an optional constructor.

Models are located in `src/models`.

Each model should extend the `AbstractModel`.
This handles the default value parsing.

A model _does not_ handle any http requests, that's what services are for.

A simple model can look like this:

```javascript
import AbstractModel from './abstractModel'
import TaskModel from './task'
import UserModel from './user'

export default class ProjectModel extends AbstractModel {
	
	constructor(data) {
		// The constructor of AbstractModel handles all the default parsing.
		super(data)
		
		// Make all tasks to task models
		this.tasks = this.tasks.map(t => {
			return new TaskModel(t)
		})
		
		this.owner = new UserModel(this.owner)
	}
	
	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			owner: UserModel,
			tasks: [],
			namespaceID: 0,
			
			created: 0,
			updated: 0,
		}
	}
}
```

#### Default values

The `defaults()` functions provides all default values.
The `AbstractModel` constructor will take all the data provided to it, and fill any non-existent,
`undefined` or `null` value with the default provided by the function.

#### Constructor

The `AbstractModel` constructor handles all the default value parsing.
In your model, the constructor can do additional parsing, like making js date object from unix timestamps
or parsing the contents of a child-array into a model.

If the model does nothing like this, you don't need to define a constructor at all.
The parent will handle it all.

#### Access to the data

After initializing a model, it is possible to access all properties via `model.property`.
To make sure the property actually exists, provide it as a default.
