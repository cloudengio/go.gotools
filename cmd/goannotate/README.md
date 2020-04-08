# goannotate

`goannotate` provides a configurable and extensible set of `annotators` that
can be used to add/remove statements from large bodies of go source code.
The original use case was to add logging to a set of APIs (defined as
interfaces and functions) to record the entry and exit from those APIs calls.
For the API components defined as interfaces it is the implementations of
those APIs that are annotated. These logging calls need to be generated in a
type and context aware manner so that they may capture the arguments and results
to the API calls.

The annotators themselves must be compiled into the `goannotate` binary (it is
possible to add new types of annotator as described [below](#adding-new-annotators))
and are configured via a yaml file. Each annotator type must be configured via
the config file and each configuration is named. This allows the configuration
file to record multiple uses of each annotator in a clean and safe manner.

## Adding new annotators

New annotators types can be added by registering them with the 
`cloudeng.io/go/cmd/goannotate/annotators` package's `Register` method.
All annotators must implement `annotators.T`. The `annotators.MustDescribe`
method provides a convenient means of documenting the configuration fields
and displaying their values. `goannotate --list` will list all available
annotator configurations.

New annotator configurations can be added directly in the config name which
each configuration requiring a unique name.
