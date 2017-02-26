# ero

Ero is a [Go](https://golang.org) binary diff checker between local VCL files and those stored in a [Fastly CDN](https://www.fastly.com/) account

The purpose of this binary is to allow you to easily check whether the version of a set of vcl files for a specific service 'version' (in Fastly terminology) is actually what you have in a local repo.

Typically I will be working within a 'staging' service environment (this is where we test out configuration before applying it to our production environment). Lots of different engineers 'borrow' the staging environment so they can test out their files, but they don't necessarily put `master` back. Meaning I don't know what's changed in comparison to the branch I happen to be working on.

This tool allows me to quickly verify which vcl files I need to update, as apposed to me blindly uploading 10+ separate vcl files via the Fastly UI.

> ero is "difference" in Finnish

## Installation

```bash
go get github.com/integralist/ero
```

Alternatively you can [download a pre-compiled binary](#)

> Once downloaded, ensure the binary is copied into your local bin path so it can be executed from any where (e.g. `cp ./ero.darwin /usr/local/bin/ero`)

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
        your service id (default: FASTLY_SERVICE_ID) (default "FASTLY_SERVICE_ID")
  -skip string
        regex for skipping vcl directories (default "^____")
  -token string
        your fastly api token (default: FASTLY_API_TOKEN) (default "FASTLY_API_TOKEN")
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
