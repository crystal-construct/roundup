# roundup
Query rancher-metadata by label, and return the first, or all of the values as a single value, csv, or a list.

Usage: 
```
roundup [options] {hosts|stacks|services|containers} predicate value-name 
```
Example:
  Fetch first container name that has the label "arangodb_cluster_name" with the value "cluster1":
```
roundup -first containers arangodb_cluster_name=cluster1 name 
```
Options:
```
-csv       Return as comma separated values
-first     Select first entry only
```
