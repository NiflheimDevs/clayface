# Lesson 3 --- HCL Basics

## Learning Objectives

By the end of this lesson you should be able to:

-   Read and write basic HCL (HashiCorp Configuration Language)
-   Understand blocks, arguments, and expressions
-   Recognize Terraform's common configuration structure
-   Write your first valid Terraform configuration

------------------------------------------------------------------------

# What is HCL?

Terraform configurations are written in **HCL (HashiCorp Configuration
Language)**.

HCL is designed to be:

-   Human-readable
-   Machine-readable
-   Declarative
-   Easy to version-control

Unlike JSON or XML, HCL is concise while remaining structured.

------------------------------------------------------------------------

# Anatomy of a Configuration

Example:

``` hcl
resource "docker_container" "nginx" {
  image = "nginx:latest"
  name  = "web"
}
```

Let's break it down.

    resource          ← Block type

    "docker_container" ← Resource type

    "nginx"            ← Local resource name

    {
        image = ...
        name  = ...
    }

Everything in Terraform is built from **blocks**.

------------------------------------------------------------------------

# Blocks

A block has this general form:

``` hcl
block_type "label1" "label2" {

}
```

Examples include:

``` hcl
terraform { }

provider "aws" { }

resource "aws_instance" "web" { }

variable "region" { }

output "ip" { }

module "network" { }
```

Not every block has two labels; the number depends on the block type.

------------------------------------------------------------------------

# Arguments

Inside blocks are **arguments**.

``` hcl
name = "web01"

memory = 4096

cpu = 2
```

General syntax:

    argument = value

------------------------------------------------------------------------

# Values

Terraform supports several value types.

## String

``` hcl
name = "server01"
```

------------------------------------------------------------------------

## Number

``` hcl
memory = 4096
```

------------------------------------------------------------------------

## Boolean

``` hcl
enabled = true
```

------------------------------------------------------------------------

## List

``` hcl
ports = [22, 80, 443]
```

------------------------------------------------------------------------

## Map (Object)

``` hcl
tags = {
  environment = "lab"
  owner       = "student"
}
```

------------------------------------------------------------------------

# Comments

Single-line:

``` hcl
# comment

// comment
```

Multi-line:

``` hcl
/*
comment
*/
```

------------------------------------------------------------------------

# Expressions

Terraform can compute values.

``` hcl
memory = 2 * 2048
```

or

``` hcl
name = "server-${1}"
```

Expressions become much more powerful once variables and functions are
introduced.

------------------------------------------------------------------------

# Formatting

Terraform includes a built-in formatter.

``` bash
terraform fmt
```

Always run it before committing code.

Benefits:

-   Consistent formatting
-   Easier code reviews
-   Reduced merge conflicts

------------------------------------------------------------------------

# Validation

Check syntax without changing infrastructure.

``` bash
terraform validate
```

This catches many configuration errors early.

------------------------------------------------------------------------

# A Minimal Configuration

``` hcl
terraform {
  required_version = ">= 1.5"
}
```

Even though it creates nothing, this is a valid Terraform configuration.

------------------------------------------------------------------------

### Confusing Block Labels

``` hcl
resource "aws_instance" "web" {

}
```

The first label is the **resource type**.

The second is Terraform's local name for that resource.

------------------------------------------------------------------------

# Summary

The most important ideas from this lesson:

-   HCL is Terraform's configuration language.
-   Everything is organized into blocks.
-   Blocks contain arguments.
-   Arguments assign values using `=`.
-   Terraform configurations are declarative.

------------------------------------------------------------------------

# Exercises

1.  Identify the block type in:

``` hcl
variable "region" {
  default = "us-east-1"
}
```

2.  Which of these are valid values?

-   `"hello"`
-   `42`
-   `true`
-   `[1,2,3]`
-   `{ key = "value" }`

3.  What command formats Terraform files?

4.  What command validates Terraform syntax?

------------------------------------------------------------------------

# Next Lesson

**Lesson 4 --- Providers**

You'll learn how Terraform downloads provider plugins, authenticates
with platforms, and why providers are the bridge between Terraform Core
and your infrastructure.
