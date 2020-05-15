k8s-prom-exporter-mgr
=====================

This tool is for automatically managing prometheus exporters in kubernetes based on scraping data from APIs.  As an example, it will create exporters in kubernetes for a dynamic number of nodes in AWS based on Tags.

Usage
=====
```
--config="./config.yml"
--once
--sleeptime=60
```

License and Author
==================

* Author:: Greg Hellings (<greg@thesub.net>)


Copyright 2020, Searchspring, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
 