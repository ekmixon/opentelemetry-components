[comment]: <> (Code generated by mdatagen. DO NOT EDIT.)

# elasticsearchreceiver

## Metrics

These are the metrics available for this scraper.

| Name | Description | Unit | Type | Attributes |
| ---- | ----------- | ---- | ---- | ---------- |
| elasticsearch.cache_memory_usage | Size in bytes of the caches. | by | Gauge | <ul> <li>server_name</li> <li>cache_name</li> </ul> |
| elasticsearch.current_documents | Number of documents in the indexes on this node. | 1 | Gauge | <ul> <li>server_name</li> <li>document_type</li> </ul> |
| elasticsearch.data_nodes | Number of data nodes in the cluster. | 1 | Gauge | <ul> </ul> |
| elasticsearch.evictions | Evictions from each cache | 1 | Sum | <ul> <li>server_name</li> <li>cache_name</li> </ul> |
| elasticsearch.gc_collection | Garbage collection count. | 1 | Sum | <ul> <li>server_name</li> <li>gc_type</li> </ul> |
| elasticsearch.gc_collection_time | Garbage collection time. | ms | Sum | <ul> <li>server_name</li> <li>gc_type</li> </ul> |
| elasticsearch.http_connections | Number of open HTTP connections to this node. | 1 | Gauge | <ul> <li>server_name</li> </ul> |
| elasticsearch.memory_usage | Size in bytes of memory. | by | Gauge | <ul> <li>server_name</li> <li>memory_type</li> </ul> |
| elasticsearch.network | Number of bytes transmitted and received on the network. | 1 | Sum | <ul> <li>server_name</li> <li>direction</li> </ul> |
| elasticsearch.nodes | Number of nodes in the cluster. | 1 | Gauge | <ul> </ul> |
| elasticsearch.open_files | Number of open file descriptors held by the server process. | 1 | Gauge | <ul> <li>server_name</li> </ul> |
| elasticsearch.operation_time | Time in ms spent on operations | ms | Sum | <ul> <li>server_name</li> <li>operation</li> </ul> |
| elasticsearch.operations | Number of operations completed | 1 | Sum | <ul> <li>server_name</li> <li>operation</li> </ul> |
| elasticsearch.peak_threads | Maximum number of open threads that have been open concurrently in the server JVM process. | 1 | Gauge | <ul> <li>server_name</li> </ul> |
| elasticsearch.server_connections | Number of open network connections to the server. | 1 | Gauge | <ul> <li>server_name</li> </ul> |
| elasticsearch.shards | Number of shards | 1 | Gauge | <ul> <li>shard_type</li> </ul> |
| elasticsearch.storage_size | Size in bytes of the document storage on this node. | by | Gauge | <ul> <li>server_name</li> </ul> |
| elasticsearch.thread_pool.active | Number of active threads in the thread pool. | 1 | Gauge | <ul> <li>thread_pool_name</li> </ul> |
| elasticsearch.thread_pool.completed | Number of tasks completed by the thread pool executor. | 1 | Sum | <ul> <li>thread_pool_name</li> </ul> |
| elasticsearch.thread_pool.queue | Number of tasks in the queue for the thread pool. | 1 | Gauge | <ul> <li>thread_pool_name</li> </ul> |
| elasticsearch.thread_pool.rejected | Number of tasks rejected by the thread pool executor. | 1 | Sum | <ul> <li>thread_pool_name</li> </ul> |
| elasticsearch.thread_pool.threads | Number of threads in the pool. | 1 | Gauge | <ul> <li>thread_pool_name</li> </ul> |
| elasticsearch.threads | Number of open threads in the server JVM process. | by | Gauge | <ul> <li>server_name</li> </ul> |

## Attributes

| Name | Description |
| ---- | ----------- |
| cache_name | Type of cache |
| direction | Data direction |
| document_type | Type of document count |
| gc_type | Type of garbage collection |
| memory_type | Type of memory |
| operation | Type of operation |
| server_name | The name of the server or node the metric is based on. |
| shard_type | State of the shard |
| thread_pool_name | Thread pool name |