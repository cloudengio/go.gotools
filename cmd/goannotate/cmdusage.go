// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Usage of goannotate:
//
// goannotate provides a configurable and extensible set of annotators
// that can be used to add/remove statements from large bodies of go source code.
//
// Command line flags:
//   -annotation string
//     	annotation to be applied
//   -config string
//     	yaml configuration file (default "config.yaml")
//   -list
//     	list available annotators
//   -list-config
//     	list available annotations and their configurations
//   -verbose
//     	display verbose debug info
//   -write-dir string
//     	if set, specify an alternate directory to write modified files to, otherwise files are modified in place.
//
// Available annotators:
//
// cloudeng.io/go/cmd/goannotate/annotators.AddLogCall:
// AddLogCall is an annotator to add function calls that are intended to log entry and exit from functions. The calls will be added as the first statement in the specified function.
//   type:                name of annotator type.
//   name:                name of annotation.
//   packages:            packages to be annotated
//   concurrency:         the number of goroutines to use, zero for a sensible
//                        default.
//   interfaces:          list of interfaces whose implementations are to be annoated.
//   functions:           list of functions that are to be annotated.
//   atLeastStatements:   the number of statements that must be present in a function
//                        in order for it to be annotated.
//   noAnnotationComment: do not annotate functions that contain this comment
//   callGenerator:       the spec for the function call to be generated
//
//     Available Call Generators:
//
//     cloudeng.io/go/cmd/goannotate/annotators/functions.LogCallWithContext
//     cloudeng.io/go/cmd/goannotate/annotators/functions.SimpleLogCall
//
//     cloudeng.io/go/cmd/goannotate/annotators/functions.LogCallWithContext:
//     LogCallWithContext provides a functon call generator for generating calls to
//     functions with the following signature:
//
//       func (ctx <contextType>, functionName, format string, arguments ...interface{}) func(ctx <contextType>, format string, namedResults ...interface{})
//
//     These are invoked via defer as show below:
//
//       defer <call>(ctx, "<function-name>",  "<format>", <parameters>....)(ctx, "<format>", <results>)
//
//     The actual type of the context is determined by the ContextType configuration
//     field. The parameters and named results are captured and passed to the logging
//     call according to cloudeng.io/go/derive.ArgsForParams and ArgsForResults.
//     The logging function must return a function that is defered to capture named
//     results and log them on function exit.
//       type:         name of annotator type.
//       importPath:   import path for the logging function.
//       functionName: name of the function to be invoked.
//       contextType:  type for the context parameter and result.
//
//     cloudeng.io/go/cmd/goannotate/annotators/functions.SimpleLogCall:
//     SimpleLogCall provides a functon call generator for generating calls to
//     functions with the same signature log.Callf and fmt.Printf.
//       type:         name of annotator type.
//       importPath:   import path for the logging function.
//       functionName: name of the function to be invoked.
//       contextType:  type for the context parameter and result.
//
//
// cloudeng.io/go/cmd/goannotate/annotators.EnsureCopyrightAndLicense:
// an annotator that ensures that a copyright and license notice is
// present at the top of all files. It will not remove existing notices.
//   type:            name of annotator type.
//   name:            name of annotation.
//   packages:        packages to be annotated
//   concurrency:     the number of goroutines to use, zero for a sensible default.
//   copyright:       desired copyright notice.
//   license:         desired license notice.
//   updateCopyright: set to true to update existing copyright notice
//   updateLicense:   set to true to update existing license notice
//
// cloudeng.io/go/cmd/goannotate/annotators.RmLogCall:
// an annotator that removes instances of calls to functions.
//   type:           name of annotator type.
//   name:           name of annotation.
//   packages:       packages to be annotated
//   concurrency:    the number of goroutines to use, zero for a sensible default.
//   interfaces:     list of interfaces whose implementations are to be annoated.
//   functions:      list of functions that are to be annotated.
//   functionNameRE: the function call (regexp) to be removed
//   comment:        optional comment that must appear in the comments associated
//                   with the function call if it is to be removed.
//   deferred:       if set requires that the function to be removed must be defered.
//
//
package main
