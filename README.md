# Ero

Ero is a cli tool, built in [Go](https://golang.org), used to diff between local & remote [Fastly CDN](https://www.fastly.com/) VCL files

## Why?

Typically when modifying VCL files, I'll be working within a 'staging' environment (this would be where we test out any VCL changes _before_ applying them to our production environment). 

Lots of different engineers 'borrow' the staging environment so they can test their changes, but they don't necessarily put the `master` version back (which leaves the stage environment in an unknown, and possibly unstable, state). 

This ultimately means I don't know what's changed in comparison to the branch I happen to be working on. What I've experienced in the past is a scenario where I upload a single VCL file to stage (this would be the file I'm modifying), but things don't work as expected because another VCL file has been changed to something from another engineer's testing branch and it causes a conflict or some other odd behaviour.

The `ero` cli tool allows me to quickly verify which files I _actually_ need to update (i.e. the file on stage isn't the same as what's in `master`). Otherwise I'll be forced to blindly upload 10+ separate VCL files via the Fastly UI to ensure that stage is in a stable state for me to upload my own changes on top of.

> ero is "difference" in Finnish

## Installation

```bash
go get github.com/integralist/ero
```

## Usage

```bash
ero -help

  -debug
        show the error/diff output
  -dir string
        vcl directory to compare files against (default "VCL_DIRECTORY")
  -help
        show available flags
  -match string
        regex for matching vcl directories
  -service string
        your service id (default "FASTLY_SERVICE_ID")
  -skip string
        regex for skipping vcl directories (default "^____")
  -token string
        your fastly api token (default "FASTLY_API_TOKEN")
  -vcl-version string
        specify Fastly service 'version' to verify against
  -version
        show application version
```

Specify credentials via cli flags:

```bash
ero -service 123abc -token 456def
```

> If no flags provided, fallback to environment vars:  
> `FASTLY_SERVICE_ID` and `FASTLY_API_TOKEN`

View the error/diff output using the debug flag:

```bash
ero -debug
```

> Typically you'll not care for the output,  
> you just want to know the files didn't match up

Specify which nested directories you want to verfiy against:

```bash
ero -match 'foo|bar'
```

> Note: .git directories are automatically ignored  
> If no flag provided, we'll look for the environment var:  
> `VCL_MATCH_DIRECTORY`

Specify which nested directories you want to skip over:

```bash
ero -skip 'foo|bar'
```

> If no flag provided, we'll look for the environment var:  
> `VCL_SKIP_DIRECTORY`

## Example

Here is an example execution:

```bash
$ ero -match www -debug

No difference between the version (123) of 'ab_tests_callback' and the version found locally
        /Users/foo/code/cdn/www/fastly/ab_tests_callback.vcl

No difference between the version (123) of 'ab_tests_deliver' and the version found locally
        /Users/foo/code/cdn/www/fastly/ab_tests_deliver.vcl

No difference between the version (123) of 'blacklist' and the version found locally
        /Users/foo/code/cdn/www/fastly/blacklist.vcl

No difference between the version (123) of 'ab_tests_config' and the version found locally
        /Users/foo/code/cdn/www/fastly/ab_tests_config.vcl

No difference between the version (123) of 'set_country_cookie' and the version found locally
        /Users/foo/code/cdn/www/fastly/set_country_cookie.vcl

No difference between the version (123) of 'ab_tests_recv' and the version found locally
        /Users/foo/code/cdn/www/fastly/ab_tests_recv.vcl

There was a difference between the version (123) of 'main' and the version found locally
        /Users/foo/code/cdn/www/fastly/main.vcl

18,21c18
<   # Blacklist check
<   # Add any IP or User-Agent that should be blacklisted to the blacklist.vcl file
<   call check_ip_blacklist;
<   call check_url_blacklist;
---
>   call check_foo_blacklist;
```

## Build

I find using [Gox](https://github.com/mitchellh/gox) the simplest way to build multiple OS versions of a Golang application:

```bash
go get github.com/mitchellh/gox

gox -osarch="linux/amd64" -osarch="darwin/amd64" -osarch="windows/amd64" -output="ero.{{.OS}}"

./ero.darwin -h
```
