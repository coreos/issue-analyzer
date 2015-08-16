
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

Advanced Usage
--------------

### Use access token

Access token could help issue-analyzer fetch data much faster.

Steps:
1. generate [personal access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/) with no scope
2. save the token into file ".oauth2_token"
