# How the binder works

The binder binds all values inside the url to their respective fields in a struct. Those fields need to have a tag
"param" with the name of the url placeholder which must be the same as in routes.

Whenever one of the standard CRUD methods is invoked, this binder is called, which enables one handler method
to handle all kinds of different urls with different parameters.
