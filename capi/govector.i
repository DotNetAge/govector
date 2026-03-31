/*
 * GoVector SWIG Interface File
 * 
 * This file configures SWIG to generate bindings for Python, Java, C#, etc.
 * 
 * Usage:
 *   swig -cgo -python -py3 govector.i      # Python
 *   swig -cgo -java govector.i             # Java
 *   swig -cgo -csharp govector.i           # C#
 */

%module govector

/* Enable CGO support */
%{
#include "govector_c.h"
%}

/* Include the C header file */
%include "govector_c.h"

/* Configure Go package */
%goheader("package main")
%gopackage("main")
