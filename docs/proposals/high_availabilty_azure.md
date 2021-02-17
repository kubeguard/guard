---
title: High Availability (Azure) | Guard
description: High Availability (Azure)
menu:
  product_guard_{{ .version }}:
    identifier: ha-azure
    name: High Availability (Azure)
    parent: proposals
    weight: 25
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: proposals
---

# Overview 

AKS (Azure Kubernetes Service) and Arc clusters provides AAD AuthN & AuthZ feature where customer can access K8s clusters using their Azure identities. With this feature, admins do not need to manage different user identities for K8s clusters and can authenticate using AAD and optionally define RBAC (role-based access control) for K8s cluster on Azure (like other Azure services) which provides him single management plane for user and access management. This feature is currently in public preview.  

To achieve the above, AKS deploys ‘Guard’ open source webhook server on K8s cluster to enable AAD authentication and authorization for AAD users. Guard is deployed as a webhook server in the cluster and receives the authentication and authorization requests from the kube-apiserver. 

This document discusses the various approaches to achieve high availability of the guard webhook server. 

# Goals 

1. Achieving high availability in the guard webhook server deployed in AKS in scenarios where the user is deploying a highly available AKS cluster. 

2. Configure a highly available and consistent cache for the guard webhook server for storing and accessing the results of every RBAC decision. We need a consistent cache as there should be only one version of the rbac decision stored across the replicas. Multiple versions can lead to inconsistencies in the final RBAC decision sent to the user. 

# Design  

High Availability in AKS is currently managed by deploying multiple replicas of the kube-apiserver which is responsible for managing the workloads in the cluster and acts as the control plane of the Kubernetes cluster. The kube-apiserver usually has 2-3 replicas in AKS for high availability depending on user requirements.  

Currently in the scenario where there is only one kube-apiserver, we have enabled the kube-apiserver cache as well, now in the high availability scenario with multiple kube-apiserver we will have to disable the kube-apiserver cache as it is an in-memory non-distributed cache. Keeping it enabled will cause inconsistency with the RBAC results. 

### One Replica of the Guard webhook server (No change) 

Pros 

1. No change is required for high availability in replicas as well cache for guard. We will still need to disable the Kubernetes authorization api cache. 

Cons: 

1. As the kube-apiserver cache will be disabled, the no of requests to guard will increase thus increasing the load on just one replica. 

2. Having just one Guard pod does not meet HA requirements i.e., if that pod goes down there is no webhook server pod to receive the requests until another pod comes up. When a zone goes down, a single replica of guard in the down zone will render the cluster unusable. 

### Guard has the same number of replicas as the kube-apiserver 

This approach proposes having the same number of guard replicas as there are of the kube-apiserver in every CCP namespace. That means we will have 2-3 replicas of guard as well. 

For this approach we will need to change the way the cache is configured in guard as all the replicas need to be consistent. 

Pros 

1. We will have high availability as there will be multiple guard pods which can receive and respond to the checkaccess requests that are coming. Thus, a sudden spike of requests or a continuous flow of a high number of requests can be easily handled. 

2. As the kube-apiserver cache will be disabled, the no of requests will increase so having more than one replica of guard will be useful to handle the request load. 

Cons: 

1. Eventual consistency of cache and depending on the approach we take based on the below sections, latency might increase too. 

## Cache in Multiple Replica Guard Scenario 

Currently in guard we use an in-memory cache (Big Cache) to store the results of the RBAC decision sent by the checkaccess call. This result is cached for 5 mins. Without a cache, guard would need to send a checkaccess request to ARM for each review request that comes from kube-apiserver. This would have resulted in increased network latency and throttling at the PAS layer. Also, if there were consecutive requests for the same resource and user from kube-apiserver, guard would have to make a checkaccess call for each of those requests. 

The result is cached only for 5 mins as the RBAC of the resource for a particular user can change depending on the role assignment changes made. Hence, we do not want to store the result permanently. As PAS recommends that check access result is valid for 5 min and RP/client should do another call after 5 min. It returns ttl in check access response which is set to 5 min as of now for all responses. Hence, we are storing the cache result for 5 mins. In the multiple guard replica scenario, we would still need to have cache as the no of requests to guard will increase as the kube-apiserver cache will be disabled. 

### Objectives 

The cache needs to be consistent with all replicas as we want to have only one version of the result and the result should be invalidated in 5 mins. 

The cache lifetime can be tied to the pod lifetime as we are not looking for permanent storage. 

### Caching strategies 

#### Sidecar Pattern using Hazelcast 

Ideally, we want a low latency, fast in-memory distributed cache known as an in memory distributed cache(IMDG). But all the embedded distributed cache offerings are in JVM, hence we are going with a sidecar pattern with an embedded distributed cache. Sidecar topology brings the benefits of Embedded topology because Hazelcast scales together with the application and both containers run on the same machine. However, the application can be written in any programming language because it connects to Hazelcast members using the standard Hazelcast Client libraries which are available in Go. We investigated Redis as a sidecar as well, but it does support distributed cache across sidecars thus will not have a consistent cache across all the replicas. 

In the event that a replica goes down, the other two replicas will be able to handle the requests and cache as well since the cache is distributed across all the replicas. On the other hand, if a new replica comes up, all the other hazelcast sidecars will discover that member and cache is distributed to it as well. 

##### How does this work? 

Guard service: This is the service for the guard deployment which handles load balancing the requests between the guard replicas. 

**Guard Pod**: 

**Main guard container** - This is the main container which runs guard. This will also have the hazelcast go client which connects to the hazelcast sidecar member through localhost itself. The client is a way to connect to the hazelcast sidecar to access the cache data. The client does not know anything about how the cache is stored, cache ttl etc. We can use the hazelcast map data structure to create a map and add key, value pairs. More on hazelcast go client (GitHub - hazelcast/hazelcast-go-client: Hazelcast IMDG Go Client) 

**Hazelcast sidecar container** - This container runs the hazelcast server instance to which the guard container connects. Using the Hazelcast Kubernetes discovery plugin, it finds all the other hazelcast server members in the defined scope. The scope can be defined using the service name and namespace. Each sidecar member will search for the other hazelcast member using the Kubernetes Api forming a hazelcast cluster. From now on, all caches (embedded in all hazelcast server applications) form one distributed caching cluster. This ensures that when the guard service sends a request to one of the guard pods, the cached result is present for all the other replicas as well making it a consistent cache. Hazelcast provides Kubernetes discovery mechanism that looks for IP addresses of other members by resolving the requests against a Kubernetes Service Discovery system. It supports two different options of resolving against the discovery registry: (i) a request to the REST API, (ii) DNS Lookup against a given DNS service name. DNS lookup is used for Headless services hence it is not applicable for guard. 

With the NODE_AWARE configuration, Hazelcast creates the partition groups with respect to member attributes map’s entries that include the node information. That means backups are created in the other nodes and each node is accepted as one partition group. 

More on Hazelcast sidecar pattern: https://hazelcast.com/blog/hazelcast-sidecar-container-pattern/ 

#### Pros 

1. We get a distributed cache with low latency as the cache can be accessed by the client on localhost itself in the container. Thus, calls to get cache data will be faster than a usual client-server cache model. 

2. We do not have to maintain a separate deployment as we would have to in a client-server approach for the cache. It would be in the same deployment as guard. 

3. Simple to implement as well. 

#### Cons: 

1. As it is a distributed cache, it has eventual consistency. Though this should be okay in our scenario as the replica count would only be 2-3. In the PoC, we ran multiple requests one after the other using a script and it has been consistent instantaneously. 

### Java in Containers 

In JDK 7/8 JVM ignores cgroups and takes the overall memory limit and core count of resources from the host system and will use that value to initialize the number of default threads. So suppose, we start 10 instances of containers on a node with a core count of 64, we will see: 

10 * 64 Jit Compiler Threads 

10 * 64 Garbage Collection threads leading to a visible decrease in performance. More on this issue: https://jaxenter.com/nobody-puts-java-container-139373.html 

To resolve this, JDK 9 supports docker cgroups, docker cpu and memory limits. Hazelcast is using JDK 11 and Hazelcast Docker image respects the container memory limits. By default, hazelcast uses 80% of the container memory limit and this value can be configured as well by passing XX:MaxRAMPercentage to the JAVA_OPTS variable 

### Metrics in Hazelcast 

Metrics can be enabled by setting the PROMETHEUS_PORT variable in the container. We can scrape the metrics by adding the Prometheus annotations as well. 

## Hazelcast cache design 

Hazelcast shards are called Partitions. By default, Hazelcast has 271 partitions. Given a key, Hazelcast will serialize, hash and mod it with the number of partitions to find the partition which the key belongs to. The partitions themselves are distributed equally among the members of the cluster. Hazelcast also creates the backups of partitions and distributes them among members for redundancy. 

Hazelcast has two types of distributed objects in terms of their partitioning strategies: 

Data structures where each partition stores a part of the instance, namely partitioned data structures. 

Data structures where a single partition stores the whole instance, namely non-partitioned data structures. 

The following are the partitioned Hazelcast data structures: 

1. Map 
2. MultiMap 
3. Cache (Hazelcast JCache implementation) 
4. Event Journal 

The following are the non-partitioned Hazelcast data structures: 

1. Queue 
2. Set 
3. List 
4. Ringbuffer 

We will be mostly using the Hazelcast Map (IMap) data structure. It extends the interface java.util.concurrent.ConcurrentMap and it the distibuted implemetation of the Java Map. 

[More non-partioned data structures](https://docs.hazelcast.org/docs/3.12.11/manual/html-single/index.html#overview-of-hazelcast-distributed-objects)

#### How is the data stored?

Hazelcast partitions your map entries and their backups, and almost evenly distribute them onto all Hazelcast members. Each member carries approximately "number of map entries * 2 * 1/n" entries, where n is the number of members in the cluster. For example, if you have a member with 1000 objects to be stored in the cluster and then you start a second member, each member will both store 500 objects and back up the 500 objects in the other member.  

[More on data partitioning in Hazelcast](https://docs.hazelcast.org/docs/3.12.11/manual/html-single/index.html#data-partitioning)

Hazelcast offers AP (Availability & Partition Tolerance)and CP (Consistency and Partition Tolerance) functionality with different data structure implementations. IMap comes under AP data structures. 

For AP data structures, Hazelcast employs the combination of primary-copy and configurable lazy replication techniques.

One of the replicas is elected as the primary replica, which is responsible for performing operations on that partition. When you read or write a map entry, you transparently talk to the Hazelcast member to which primary replica of the corresponding partition is assigned. By this way, each request hits the most up-to-date version of a particular data entry in a stable cluster. Backup replicas stay in standby mode until the primary replica fails. Upon failure of the primary replica, one of the backup replicas is promoted to the primary role. 

With lazy replication, when the primary replica receives an update operation for a key, it executes the update locally and propagates it to backup replicas. It marks each update with a logical timestamp so that backups apply them in the correct order and converge to the same state with the primary. 

[More on consistency and replication](https://docs.hazelcast.org/docs/3.12.11/manual/html-single/index.html#consistency-and-replication-model)

#### Consistency in backups

To provide data safety, Hazelcast allows you to specify the number of backup copies you want to have. That way, data on a cluster member is copied onto other member(s). There is sync as well as async backups. By default backup operations ae synchronous. In this case, backup operations block operations until backups are successfully copied to backup members (or deleted from backup members in case of remove) and acknowledgements are received. Therefore, backups are updated before a put operation is completed, provided that the cluster is stable. 

By default, Hazelcast has one sync backup copy. If backup-count is set to more than 1, then each member will carry both owned entries and backup copies of other members. So for the map.get(key) call, it is possible that the calling member has a backup copy of that key. By default, map.get(key) always reads the value from the actual owner of the key for consistency. 

**Best Effort Consistency**: Due to temporary situations in the system, such as network interruption, backup replicas can miss some updates and diverge from the primary. Backup replicas can also hit VM or long GC pauses, and fall behind the primary, which is a situation called as replication lag. If a Hazelcast partition primary replica member crashes while there is a replication lag between itself and the backups, strong consistency of the data can be lost. To minimize the effect of such scenarios using an active anti-entropy solution as follows: 

Each Hazelcast member runs a periodic task in the background. 

1. For each primary replica it is assigned, it creates a summary information and sends it to the backups. 

2. Then, each backup member compares the summary information with its own data to see if it is up-to-date with the primary. 

3. If a backup member detects a missing update, it triggers the synchronization process with the primary. 

Please see https://docs.hazelcast.org/docs/3.12.11/manual/html-single/index.html#map for further details.

#### Features
1. If a member goes down, its backup replica (which holds the same data) dynamically redistributes the data, including the ownership and locks on them, to the remaining live members. As a result, there will not be any data loss. 

2. There is no single cluster primary that can be a single point of failure. Every member in the cluster has equal rights and responsibilities. No single member is superior. There is no dependency on an external 'server' or 'master'. 

#### More on IMAP
IMap.put does not acquire any lock. Each Hazelcast member has certain set of partitions, which are handled by a number of partition threads. Each partition thread owns certain number of partitions. So when a write operation arrives at a cluster member, it is picked by the partition thread that owns the partition which would be the host of the Entry object in that write operation. 

If the same member receives another write operation, which is destined for the same or other partitions owned by the same partition thread, it would not be picked until the previous write operation by the thread is complete. This prevents any potential race conditions and does not require explicit locking in default state.

## Client-Server Cache model using Redis 

#### What do we need? 

**Client side** - Golang Redis client library that can connect to the Redis server. We also need to be able to set the ttl for each key to 5 mins -  https://github.com/go-redis/redis. We will need secrets to store the Redis Host URL and Redis token secret. 

**Server side** - Redis in Kubernetes has primary-secondary architecture in the client-server model. We investigated what are the norms to deploy Redis in Kubernetes and found that Bitnami provides a Redis chart that can be deployed in two ways: 

Redis | Redis Cluster
--- | --- 
Single write point (single primary) | Multiple write points (multiple primary) 
Default Configuration  | 290 
Primary supports read-write operations and secondaries only support read operations.  | Multiple writing points as multiple primary 
In this configuration, data written on the primary is replicated on all secondaries.  | In this configuration, data is sharded.
Without Redis Sentinel enabled, the secondaries will wait until the primary node is respawned again by the Kubernetes Controller Manager. With Redis Sentinel enabled, in case the current primary crashes, the Sentinel containers will elect a new primary node. | Redis Cluster can survive partitions where the majority of the primary nodes are reachable and there is at least one reachable secondary for every primary node that is no longer reachable. If the primary node or even all the nodes are down, the cluster is automatically recovered, and new primary nodes are promoted to maintain the balance of the cluster and ensure that read/write operations continue without interruption.
This would be in every CCP. | This might be overkill if it is deployed in every CCP but can work if deployed in the underlay which has many ccp's 

## Testing Scenarios 

To be decided 

## References  

https://hazelcast.com/blog/hazelcast-sidecar-container-pattern/ 

https://hazelcast.com/blog/architectural-patterns-for-caching-microservices/ 

https://github.com/hazelcast/hazelcast-go-client 

https://github.com/hazelcast/hazelcast-kubernetes 

https://docs.hazelcast.org/docs/3.12.11/manual/html-single/index.html#map

https://docs.bitnami.com/kubernetes/infrastructure/redis/get-started/cluster-topologies/ 

https://github.com/kubernetes/examples/tree/master/staging/storage/redis 

https://docs.bitnami.com/kubernetes/infrastructure/redis-cluster/get-started/compare-bitnami-solutions/ 

https://medium.com/@inthujan/introduction-to-redis-redis-cluster-6c7760c8ebbc 

 