
issue-analyzer
=====

issue-analyzer helps to analyze github repo issue counts over time.


Installation
------------

```
go get github.com/yichengq/issue-analyzer
```


Usage
-----

run `./issue-analyzer`, which generates png files at current directory.

Flags:
```
  -end-date string
    	end date of the graph, in format 2000-Jan-01 or 2000-Jan
  -owner string
    	the owner in github (default "coreos")
  -repo string
    	the repo of the owner in github (default "etcd")
  -start-date string
    	start date of the graph, in format 2000-Jan-01 or 2000-Jan
  -token string
    	access token for github
```

Advanced Usage
--------------

### Use access token

Access token could help issue-analyzer fetch data much faster.

Steps:
1. generate [personal access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/) with no scope
2. save the token into file ".oauth2_token"
