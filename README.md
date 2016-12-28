# JSON API Check

**check-json - Use Nagios / Icinga to monitor JSON endpoints**

This plugin tests JSON API endpoints served over HTTP. It can check for the
existence of keys or do simple checks against values in the JSON response.

Command line flags have been chosen to be compatible with the common Nagios
[check_http](https://www.monitoring-plugins.org/doc/man/check_http.html) plugin.
Pattern for JSON tests from [Python JSON
Check](https://github.com/drewkerrigan/nagios-http-json). This version allows
regex test of string values and is a stand-alone binary.

Basic Usage:
```bash
  check-json --hostname=time.jsontest.com \
    --key-exists=time --key-equals=date:2016
```

Response Check Options:
```
  -e, --key-exists=    Checks existence of these keys from JSON response
  -q, --key-equals=    A regex to check the value of specific key values from
                       JSON response
  -l, --key-lte=       Check the returned value is less than this for a JSON key
  -g, --key-gte=       Check the returned value is greater than this for a JSON
                       key
  -d, --header-equals= Key=value checks for HTTP response headers (key:value)
  -s, --status=        Checks the numerical HTTP return status (eg. 200)
  -r, --regexp=        Checks the response body for a string using a regular
                       expression.
  -v, --verbose        Display extra details (eg. response bodies) for debugging
                       (false)
```

HTTP Options:
```
  -H, --hostname=      Web server to query
  -u, --uri=           URI to GET or POST (/)
  -j, --method=        HTTP method (eg. HEAD, OPTIONS, TRACE, PUT, DELETE) (GET)
  -P, --post=          Body of POST Request
  -a, --authorization= Basic HTTP auth (username:password)
  -S, --ssl            Enforce SSL (false)
  -k, --header=        Key,value pairs to add as headers in HTTP request
                       (name:value format)
```

Help Options:
```bash
  -h, --help           Show this help message
```

## Installation

Ensure you have go installed. On a mac you can use [homebrew](http://brew.sh):
```bash
brew install go
```

Ensure your environment is setup correctly by putting vars in your `.bashrc`
file:

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```
Your GOHOME can be anywhere you choose.

Building is more painful than normal because it is not a public repo.

```
go build github.com/werrett/check-json
```

Now Check that you have it installed OK:

```bash
check-json --help

Usage:
  check-json [OPTIONS]

Application Options:
  ...
```
## Example Commands

Simple JSON key exists and regex against key value:

```bash
check-json --hostname=time.jsontest.com  \
  --key-exists=time --key-equals=date:2016 --verbose
```

Simple check using SSL:

```bash
check-json --ssl --hostname=api.foursquare.com \
   --uri='/v2/venues/4e37bb6aa809a0c63b3882e8?client_id=AA...&client_secret=XX...&v=20150313' \
   --key-exists=response
```

Check adding an authentication header:

```bash
check-json --hostname=company.clearbit.co \
  --uri=/v1/companies/domain/clearbit.co --ssl \
  --header="Authorization:Bearer sk_..." \
  --key-equals="name:Clearbit" --verbose
```


## Todo

 - Nagios performance data
 - Checks on load times (warning / critical)
 - Integration tests
