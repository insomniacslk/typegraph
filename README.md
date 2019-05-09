# typegraph

typegraph generates GraphViz DOT files from Go source code.

It works by parsing the AST of the Go source files passed as input, traversing
all the declared types, and generating a digraph of the connections among types.

An example of this program running on its own source code:

![typegraph](res/graph.png)

# Bugs?

So many. The prototype has been developed in a hour, with no prior experience on AST.
