/*
Usage:
  go-post [-diff] [-r fixname,...] [-force fixname,...] [path ...]

Flags:
  -diff
    	display diffs instead of rewriting files
  -force string
    	force these fixes to run even if the code looks updated
  -r string
    	restrict the rewrites to this comma-separated list

Available rewrites are:

* assignbinop

Replace "x = x + z" with "x += z".

* deadassign

Remove "x = x" assignments.

* forloop

Add initialization and post-statements to for-loops.

* localid

Replace the use of local variable IDs with their definition.

* mainret

Replace return statements with calls to os.Exit in the "main" function.

* unresolved

Replace assignment statements with declare and initialize statements at the first occurance of an unresolved identifier.
*/
package main
