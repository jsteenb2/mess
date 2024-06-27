# How to Inherit a Mess

Welcome all! This git repo outlines most everything we're going to cover in the
workshop. There will be some quips here and there that may not be captured here,
but by and large this will be the entirety. The below walks through the codebase
as it was developed. For the best possible outcomes, follow each step in the
order they are given.

Make sure you have the following on your machine to get started:
* go [1.22+](https://go.dev/dl/)
* text editor/IDE
* pull this [repo](https://github.com/jsteenb2/mess)
* have this Readme open in github (very handy for navigation)

Most of the Questions listed below are aimed at self study(i.e. may not have time in
the in person workshop). Feel free to create a branch from a commit and make changes
to your hearts content. It is not recommended you futz with the git history of the
main branch though, as that will break all the navigation listed below.

The contents of the workshop are ordered by the git repository's commit message. This
README is generated from those commits. From this we take our first lesson of the
workshop:

> Write meaningful commits

If you play your cards right, you can do a whole lot with your commit history. Use a
commit to add any tribal knowledge that isn't clear from the code/comments. At the
very least, leave a bread crumb in the commit that outlines how and why the decision(s)
were made. Additionally, when you choose a specification like [conventional commits](https://www.conventionalcommits.org),
you have some structure to tie into. This README is generated with a simple
[template engine](https://github.com/jsteenb2/gitempl). Adding a little structure
to the commit log made it all possible.

---

## Table of Contents

1. [Introduction](#1-introduction-top)
2. [Add implementation of allsrv server](#2-add-implementation-of-allsrv-server-top)
3. [Add high-level implementation notes](#3-add-high-level-implementation-notes-top)
4. [Add tests for create foo endpoint](#4-add-tests-for-create-foo-endpoint-top)
5. [Inject id generation fn](#5-inject-id-generation-fn-top)
6. [Panic trying to add a test for the read API](#6-panic-trying-to-add-a-test-for-the-read-api-top)
7. [Replace `http.DefaultServeMux` with isolated `*http.ServeMux` dependency](#7-replace-httpdefaultservemux-with-isolated-httpservemux-dependency-top)
8. [Adding test for update foo API](#8-adding-test-for-update-foo-api-top)
9. [Add test for the delete foo API](#9-add-test-for-the-delete-foo-api-top)
10. [Fixup false not found error in delete foo API](#10-fixup-false-not-found-error-in-delete-foo-api-top)
11. [Add tests for unauthorized access](#11-add-tests-for-unauthorized-access-top)
12. [DRY auth with a basic auth middleware](#12-dry-auth-with-a-basic-auth-middleware-top)
13. [Inject authorization mechanism to decouple it from server](#13-inject-authorization-mechanism-to-decouple-it-from-server-top)
14. [Refactor DB interface out of in-mem database type](#14-refactor-db-interface-out-of-in-mem-database-type-top)
15. [Add database observer for DB metrics](#15-add-database-observer-for-db-metrics-top)
16. [Add tracing to the database observer](#16-add-tracing-to-the-database-observer-top)
17. [Add http server observability](#17-add-http-server-observability-top)
18. [Add tests for the in-mem db](#18-add-tests-for-the-in-mem-db-top)
19. [Serialize access for in-mem value access to rm race condition](#19-serialize-access-for-in-mem-value-access-to-rm-race-condition-top)
20. [Add structured errors](#20-add-structured-errors-top)
21. [A note on selling change](#21-a-note-on-selling-change-top)
22. [Add v2 Server create API](#22-add-v2-server-create-api-top)
23. [Refactor v2 tests with table tests](#23-refactor-v2-tests-with-table-tests-top)
24. [Extend ServerV2 API with read/update/delete](#24-extend-serverv2-api-with-readupdatedelete-top)
25. [Add simple server daemon for allsrv](#25-add-simple-server-daemon-for-allsrv-top)
26. [Add deprecation headers to the v1 endpoints](#26-add-deprecation-headers-to-the-v1-endpoints-top)
27. [Add db test suite](#27-add-db-test-suite-top)
28. [Add sqlite db implementation and test suite](#28-add-sqlite-db-implementation-and-test-suite-top)
29. [Add the service layer to consolidate all domain logic](#29-add-the-service-layer-to-consolidate-all-domain-logic-top)
30. [Provide test suite for SVC behavior and fill gaps in behavior](#30-provide-test-suite-for-svc-behavior-and-fill-gaps-in-behavior-top)
31. [Add http client and fixup missing pieces](#31-add-http-client-and-fixup-missing-pieces-top)
32. [Add allsrvc CLI companion](#32-add-allsrvc-cli-companion-top)
33. [Add pprof routes](#33-add-pprof-routes-top)
34. [Add github.com/jsteenb2/errors module to improve error handling](#34-add-githubcomjsteenb2errors-module-to-improve-error-handling-top)
35. [Add github.com/jsteenb2/allsrvc SDK module](#35-add-githubcomjsteenb2allsrvc-sdk-module-top)
36. [Add slimmed down wild-workouts-go-ddd-example repo](#36-add-slimmed-down-wild-workouts-go-ddd-example-repo-top)
37. [Add thoughts on wild-workouts Clean architecture](#37-add-thoughts-on-wild-workouts-clean-architecture-top)
38. [References](#references-top)
39. [Suggested Resources](#suggested-resources-top)

---

### 1. Introduction [[top]](#how-to-inherit-a-mess)

This workshop walks through a very familiar situation we've all found ourselves in at
some point in our careers. It starts something like this:

> I just accepted a job with $COMPANY! I can't wait to get started!

It's an exciting time. Endless possibilities await you. The days leading up to this
new position are filled with anticipation and excitement. Then finally, the time comes
for you to start the new role. The anticipation for this day carries you through the
onboarding slog. You sum up your first week:

> Wow! This is so exciting! So much to learn.... I'll be drinking from the firehose for a while!

Fast forward six months to a year and somehow the mood has soured. What was once amazement has
morphed into a laundry list of grievances:

 1. Lack of clear intention behind the system design/codebase
 2. CI process that is counted in 10min increments
 3. Tests.... where are they?
 4. gorm v1
 5. Found the tests.... and they're completely coupled to the implementation
 6. Ughhhh...

We've all found ourselves in this visceral quagmire at some point in our career. We have
options when we hit this point:

1. We polish off our resume and look for the green grass elsewhere
    * I don't blame anyone for taking this route
2. Acclimate yourself to the mess
    * Effectively give up hope that change is a possibility and just mosey along. Very common when golden handcuffs are in play
3. Be an agent of change!

⚠️ Disclaimer: you can assume all three options simultaneously, they are not exclusive by any means.

Thi workshop is primarily aimed at the person putting on the option 3 hat. Though I will warn you
it can be an exhausting, unforgiving, and unrewarding endeavor... When you're able to get
buy in from your team and your management, and you're convinced the juice is worth the squeeze...
The only thing left is to equip yourselves with the tools to get the job done.

Welcome to "How to Inherit a Mess"!

### 2. Add implementation of allsrv server [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/643f58a1361bdaa3515866b73125e0558ac0efc0)**

```shell
git checkout 643f58a
```

<details>
<summary>Files impacted</summary>

| File                                 | Count   | Diff                                                                                   |
|--------------------------------------|---------|----------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **166** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [go.mod](go.mod)                     | **5**   | <span style="color:green">+++++</span>                                                 |
| [go.sum](go.sum)                     | **2**   | <span style="color:green">++</span>                                                    |


</details>

This server represents a server with minimum abstraction. It takes the
popular convention from the Matt Ryer post and implements it in an intensely
terse manner.

Explore this allsrv pkg and answer the following questions:

#### Questions

* [ ] What stands out?
* [ ] What do you find favorable?
* [ ] What do you find nasueating?


<details><summary>References</summary>

1. [How I write HTTP Services After 8 Years - Matt Ryer](https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)

</details>

### 3. Add high-level implementation notes [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/733884c017ce35aca773df96c63cebc4cde316cb)**

```shell
git checkout 733884c
```

<details>
<summary>Files impacted</summary>

| File                                 | Count  | Diff                                                                                                                   |
|--------------------------------------|--------|------------------------------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **84** | <span style="color:green">++++++++++++++++++++++++++++++++++++++++</span><span style="color:red">--------------</span> |
| [go.mod](go.mod)                     | **2**  | <span style="color:green">+</span><span style="color:red">-</span>                                                     |


</details>

These are some of my thoughts around this starter project. In this
workshop, we'll look at a couple different examples of codebases.
This `allsrv` pkg represents one extreme: the wildly understructured
variety. Later we'll explore the other extreme, the intensely
overstructured kind.

<details><summary>Spoiler</summary>

<h4>Praises</h4>

1. minimal public API
2. simple to read
3. minimal indirection/obvious code
4. is trivial in scope

<h4>Concerns</h4>

1. the server depends on a hard type, coupling to the exact inmem db
	* what happens if we want a different db?
2. auth is copy-pasted in each handler
	* what happens if we forget that copy pasta?
3. auth is hardcoded to basic auth
	* what happens if we want to adapt some other means of auth?
4. router being used is the GLOBAL http.DefaultServeMux
	* should avoid globals
	* what happens if you have multiple servers in this go module who reference default serve mux?
5. no tests
	* how do we ensure things work?
	* how do we know what is intended by the current implementation?
6. http/db are coupled to the same type
	* what happens when the concerns diverge? aka http wants a shape the db does not? (note: it happens A LOT)
7. Server only works with HTTP
	* what happens when we want to support grpc? thrift? other protocol?
	* this setup often leads to copy pasta/weak abstractions that tend to leak
8. Errors are opaque and limited
9. API is very bare bones
	* there is nothing actionable, so how does the consumer know to handle the error?
	* if the APIs evolve, how does the consumer distinguish between old and new?
10. Observability....
11. hard coding UUID generation into db
12. possible race conditions in inmem store

</details>

I want to make sure you don't get the wrong impression. We're not here to learn
how to redesign *all the things*. Rather, we're here to increase the number of
tools in our toolbelt to deal with a legacy system. Replacing a legacy system is only one
of many possible courses of action. In the event the legacy system is unsalvageable,
a rewrite may be the only course of action. The key is understanding the problem space
in its entirety.

Our key metric in navigating this mess, is understanding how much non-value added
work is being added because of the limitations of the legacy system. Often times
there is a wishlist of asks from product managers and customers alike that are
*"impossible"* because of some warts on the legacy design. As we work through
this repo, try to envision different scenarios you've found yourself in. We'll
cover a broad set of topics and pair this with the means to remedy the situation.
Here's what we're going to cover:

1. Understanding the existing system
    * What do our tests tell us?
    * What gaps do we have to fill in our tests to inform us of different aspects of the legacy design?
    * What's the on call experience like?
    * What's the onboarding experience like?
    * How large are typical PRs?
    * How comfortable do engineers feel contributing changes?
    * How much of the team contributes changes?
2. Understand the external constraints imposed on the system that are out of your control
    * Verify the constraints are truly required
        * Most common bad assumption I see is automagik replication... we say this like its a must, but rarely do we understand the full picture. Its often unnecessary and undesirable from a user perspective.
    * The key is understanding the way in which the parts of the system interact
3. Understanding the cost of *abstraction*
    * Clean, Hexagonal, Domain Driven Design, etc... these all provide value, but are they worth the cost?
    * *DRY* up that code?
    * Microservice all the things `:table_flip:`
    * API should mimic the entire database `:double_table_flip:`
    * Follow what Johnny shared, because workshops dont' lie... dogma wins you nothing
4. Understand the customer/user's experience
5. Codify tribal knowledge
    * CLIs are your friend. They can be used to codify a ton of tribal knowledge. Done right they can be extremely user friendly as well. Shell completions ftw!
6. Much more `:yaaaaaaas:`

My goal is each person taking this workshop will walk away with a few nuggets
of wisdom that will improve their day to day.



### 4. Add tests for create foo endpoint [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/e9374a049f258ea8b7dacdd8a377039eb5bf0b0e)**

```shell
git checkout e9374a0
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                   |
|------------------------------------------------|--------|----------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go)           | **15** | <span style="color:green">++++++++</span><span style="color:red">-------</span>        |
| [allsrv/server_test.go](allsrv/server_test.go) | **66** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [go.mod](go.mod)                               | **11** | <span style="color:green">++++++++++</span><span style="color:red">-</span>            |
| [go.sum](go.sum)                               | **10** | <span style="color:green">++++++++++</span>                                            |


</details>

We are up against the wall with the UUID generation
hardcoded in the db create. This makes the test
non-deterministic.

#### Questions

* [ ] How do we make UUID generation determinstic?


### 5. Inject id generation fn [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/1a50c7e6ebf775d58893a491b5001a9597d56c0f)**

```shell
git checkout 1a50c7e
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                                     |
|------------------------------------------------|--------|----------------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go)           | **40** | <span style="color:green">+++++++++++++++++++++++++</span><span style="color:red">---------------</span> |
| [allsrv/server_test.go](allsrv/server_test.go) | **9**  | <span style="color:green">++++</span><span style="color:red">-----</span>                                |


</details>

Motivation for this change:

  1. we want determinism in tests, this allows for that
  2. we don't want the db owning the business logic of what an ID looks like
    a) if we switch out dbs, each db has to make sure its ID gen aligns... youch

Our tests are now fairly easy to follow. At this time, our `Server` is really
the owner of the id gen business logic. This is ok for now, but for more
complex scenarios, this can pose a serious problem.



### 6. Panic trying to add a test for the read API [[top]](#how-to-inherit-a-mess)

* pkg: **allsvr**
* commit: **[github](https://github.com/jsteenb2/mess/commit/efa38f407a88e79692e01ec87206bdce8f6d9267)**

```shell
git checkout efa38f4
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                                             |
|------------------------------------------------|--------|------------------------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go)           | **4**  | <span style="color:green">++</span><span style="color:red">--</span>                                             |
| [allsrv/server_test.go](allsrv/server_test.go) | **48** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++</span><span style="color:red">---------</span> |


</details>

What's going on here? Take a moment to reflect on this. What is the simplest
possible fix here?



### 7. Replace `http.DefaultServeMux` with isolated `*http.ServeMux` dependency [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/526c39c1df8c767c06e2e5189027a4644b53c627)**

```shell
git checkout 526c39c
```

<details>
<summary>Files impacted</summary>

| File                                 | Count  | Diff                                                                                                           |
|--------------------------------------|--------|----------------------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **46** | <span style="color:green">++++++++++++++++++++++++</span><span style="color:red">----------------------</span> |


</details>

This resolves the panic adding routes with the same pattern multiple
times. Now each `Server`, has its own `*http.ServeMux`. Now tests run independent
of one another and we avoid the pain of GLOBALS!

The tests should now pass :-)



### 8. Adding test for update foo API [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/1e3490587a9a8b204bfe4b3b5c41193ba5eb09ab)**

```shell
git checkout 1e34905
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                         |
|------------------------------------------------|--------|--------------------------------------------------------------|
| [allsrv/server_test.go](allsrv/server_test.go) | **27** | <span style="color:green">+++++++++++++++++++++++++++</span> |


</details>

The key takeaway here is the response is a nothing burger. We can only ever
test the status code atm. These leaves a lot to be desired. As a service
evolves to cover a broader domain, these types tend to grow. We don't see
anything in an empty response however.

At the moment we're kind of stuck though. We don't have a means to add a
new API that is distinguished from the existing. We'll come back to this.



### 9. Add test for the delete foo API [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/6a573b27a75d66a9349e7d65e66ed169d55a8003)**

```shell
git checkout 6a573b2
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                    |
|------------------------------------------------|--------|---------------------------------------------------------|
| [allsrv/server_test.go](allsrv/server_test.go) | **22** | <span style="color:green">++++++++++++++++++++++</span> |


</details>

This test currently fails, but it's not entirely obvious as to why. Fix
the bug and make these tests pass.



### 10. Fixup false not found error in delete foo API [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/ec22efa0f49cdb8bc0daab6f5add7a62affec467)**

```shell
git checkout ec22efa
```

<details>
<summary>Files impacted</summary>

| File                                 | Count | Diff                                |
|--------------------------------------|-------|-------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **2** | <span style="color:green">++</span> |


</details>

The tests should now pass. Without the test, this isn't immediately obvious.
Point in case, I didn't realize I missed this until I wrote the test in the
previous commit. That's why you won't find it in the list of concerns!



### 11. Add tests for unauthorized access [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/79af2cb1c69471c59dfb681cdcf9f4ccf593eaac)**

```shell
git checkout 79af2cb
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                   |
|------------------------------------------------|--------|----------------------------------------------------------------------------------------|
| [allsrv/server_test.go](allsrv/server_test.go) | **63** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |


</details>

Filling in some tests gaps. With these in place, we can now address
concern 2), the duplication of auth everywhere. Take a crack at
DRYing up the basic auth integration.



### 12. DRY auth with a basic auth middleware [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/5030afa460766e21db28f73b530d950ad7e21c21)**

```shell
git checkout 5030afa
```

<details>
<summary>Files impacted</summary>

| File                                 | Count  | Diff                                                                                                               |
|--------------------------------------|--------|--------------------------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **50** | <span style="color:green">+++++++++++++++++++++</span><span style="color:red">-----------------------------</span> |


</details>

This removes the duplication of code seen throughout our handlers.
With our tests in place, we can refactor this safely.



### 13. Inject authorization mechanism to decouple it from server [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/1a259d675f8976c6a74bc3aa23752ae60ecd6bfb)**

```shell
git checkout 1a259d6
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                              |
|------------------------------------------------|--------|---------------------------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go)           | **33** | <span style="color:green">++++++++++++++++++++++</span><span style="color:red">-----------</span> |
| [allsrv/server_test.go](allsrv/server_test.go) | **23** | <span style="color:green">+++++++++++++</span><span style="color:red">----------</span>           |


</details>

This does a few things:

  1) Allows us to control the auth at setup without having to update
     the implementation of the server endpoints. We've effectively decoupled
     our auth from the server, which gives us freedom to adapt to future
     asks.
  2) The injection is using a `middleware` function. This could be
     an interface as well. Its totally up to the developer/team. Sometimes
     an interface is more useful.
  3) We have the freedom to ignore auth in tests if we so desire. This
     can be useful if your auth setup is non-trivial and involves a good
     bit of complexity.



### 14. Refactor DB interface out of in-mem database type [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/c91f0920a056d4f6732ba802b34c8f940020a08f)**

```shell
git checkout c91f092
```

<details>
<summary>Files impacted</summary>

| File                                     | Count  | Diff                                                                                                                  |
|------------------------------------------|--------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/db_inmem.go](allsrv/db_inmem.go) | **51** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                  |
| [allsrv/server.go](allsrv/server.go)     | **70** | <span style="color:green">+++++++++++++</span><span style="color:red">----------------------------------------</span> |


</details>

This paves the way for more interesting additions. With this DB
interface in place, can you add metrics for the datastore without
futzing with the in-mem database implmementation? Add/update tests
to verify the db still exhibits the same behavior.



### 15. Add database observer for DB metrics [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/f299727847bf6f82bbf5d60f8a5f79147aa04608)**

```shell
git checkout f299727
```

<details>
<summary>Files impacted</summary>

| File                                           | Count   | Diff                                                                                                                  |
|------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/observe_db.go](allsrv/observe_db.go)   | **62**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/server.go](allsrv/server.go)           | **3**   | <span style="color:green">+++</span>                                                                                  |
| [allsrv/server_test.go](allsrv/server_test.go) | **18**  | <span style="color:green">++++++++++++++</span><span style="color:red">----</span>                                    |
| [go.mod](go.mod)                               | **6**   | <span style="color:green">++++++</span>                                                                               |
| [go.sum](go.sum)                               | **104** | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++++</span><span style="color:red">-</span> |


</details>

Often times, metrics and tracing are left out, leaving the service
owners blind. When we add metrics, with defined patterns, we're able to
build out robust metrics and tracing dashboards.

Additionally, with the observer we get a warning from the compiler
that we forgot to add observability concerns for any new behavior/method
added to the database(s).

Now, how about adding opentracing spans to the observer?



### 16. Add tracing to the database observer [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/4f89d8c5e8ff06b25bbfeedffb2225bc6eb1a163)**

```shell
git checkout 4f89d8c
```

<details>
<summary>Files impacted</summary>

| File                                           | Count  | Diff                                                                                           |
|------------------------------------------------|--------|------------------------------------------------------------------------------------------------|
| [allsrv/db_inmem.go](allsrv/db_inmem.go)       | **9**  | <span style="color:green">+++++</span><span style="color:red">----</span>                      |
| [allsrv/observe_db.go](allsrv/observe_db.go)   | **30** | <span style="color:green">++++++++++++++++++++++</span><span style="color:red">--------</span> |
| [allsrv/server.go](allsrv/server.go)           | **17** | <span style="color:green">+++++++++</span><span style="color:red">--------</span>              |
| [allsrv/server_test.go](allsrv/server_test.go) | **9**  | <span style="color:green">+++++</span><span style="color:red">----</span>                      |
| [go.mod](go.mod)                               | **1**  | <span style="color:green">+</span>                                                             |
| [go.sum](go.sum)                               | **2**  | <span style="color:green">++</span>                                                            |


</details>

The observer is updated with additional observability concerns. This
does violate single responsibility principles, but in this case it
encapsulates well. In my experience, it's rare that you have
metrics and tracing concerns that are required, and you only want one
or the other. Having a single observer for these makes it fairly simple.

Now that these are in place, add observability for metrics/tracing to
the http server.



### 17. Add http server observability [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/edfa256311e7c47ed28bb5be9d36f06ddde9b589)**

```shell
git checkout edfa256
```

<details>
<summary>Files impacted</summary>

| File                                                               | Count   | Diff                                                                                     |
|--------------------------------------------------------------------|---------|------------------------------------------------------------------------------------------|
| [allsrv/observer_http_handler.go](allsrv/observer_http_handler.go) | **110** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>   |
| [allsrv/server.go](allsrv/server.go)                               | **4**   | <span style="color:green">++</span><span style="color:red">--</span>                     |
| [allsrv/server_test.go](allsrv/server_test.go)                     | **24**  | <span style="color:green">++++++++++++++++</span><span style="color:red">--------</span> |


</details>

The HTTP server now has quantification for different metrics important
to an HTTP server. The basis of our observability is now in place. We
can now create dashboards/insights to understand the deployed service.

One thing to note here is we have not touched on logging just yet. Good
logging is inherently coupled to good error handling. We'll wait until
we have a better handle of our error handling before proceeding.



### 18. Add tests for the in-mem db [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/ecfd46362259c63ac7cfbee79e89223772b1e68f)**

```shell
git checkout ecfd463
```

<details>
<summary>Files impacted</summary>

| File                                               | Count   | Diff                                                                                   |
|----------------------------------------------------|---------|----------------------------------------------------------------------------------------|
| [allsrv/db_inmem_test.go](allsrv/db_inmem_test.go) | **154** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |


</details>

This helps us close the gap in our testing. This time we're putting the
in-mem db under test. This is partially under test via the server tests,
but we have limited visibility into the stack. Once we have the basics
in place, we can start to ask more interesting quetsions of our system.

Try to create a test that will trigger the race condition in the in-mem
operations for each destructive operation? Hint: use the `-race` flag:

```shell
go test -race
```



### 19. Serialize access for in-mem value access to rm race condition [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/584430174d957f48ae02be81484908c204ef8a28)**

```shell
git checkout 5844301
```

<details>
<summary>Files impacted</summary>

| File                                               | Count   | Diff                                                                                   |
|----------------------------------------------------|---------|----------------------------------------------------------------------------------------|
| [allsrv/db_inmem.go](allsrv/db_inmem.go)           | **16**  | <span style="color:green">+++++++++++++++</span><span style="color:red">-</span>       |
| [allsrv/db_inmem_test.go](allsrv/db_inmem_test.go) | **119** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [allsrv/server.go](allsrv/server.go)               | **4**   | <span style="color:green">++</span><span style="color:red">--</span>                   |


</details>

This is pretty straight forward. We just add the `-race` flag to our `go test`
invocation after we add some tests that access the db concurrently. There
are many more test cases to add for this specific instance, but the point
is made with the existing one.

We're starting to get somewhat comfortable with our existing test suite. If
this were a real world service, you'd have tests at a higher level and more
integration and end to end testing. These tests are slow but are the most
valuable as they are validating more integration points! Unit tests can easily create a
false sense of safety. Especially when unit testing that is MOCKING ALL
THE THINGS!

Note, we can use other mechanisms to address the race condition. However,
it gets complex... FAST. Katherine Cox-Buday's "Concurrency in Go" is a
wonderful deep dive on the subject :-):



<details><summary>References</summary>

2. [Concurrency in Go - Katherine Cox-Buday](https://katherine.cox-buday.com/concurrency-in-go/)

</details>

### 20. Add structured errors [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/98c8963c41270580481e9f63f9576ec501dfe6b8)**

```shell
git checkout 98c8963
```

<details>
<summary>Files impacted</summary>

| File                                               | Count  | Diff                                                                       |
|----------------------------------------------------|--------|----------------------------------------------------------------------------|
| [allsrv/db_inmem.go](allsrv/db_inmem.go)           | **9**  | <span style="color:green">++++</span><span style="color:red">-----</span>  |
| [allsrv/db_inmem_test.go](allsrv/db_inmem_test.go) | **10** | <span style="color:green">+++++</span><span style="color:red">-----</span> |
| [allsrv/errors.go](allsrv/errors.go)               | **38** | <span style="color:green">++++++++++++++++++++++++++++++++++++++</span>    |
| [allsrv/server.go](allsrv/server.go)               | **2**  | <span style="color:green">+</span><span style="color:red">-</span>         |


</details>

Structured errors are so incredibly helpful for service owners. With
structured errors, you're able to enrich the context of your failures
without having to add DEBUG log after DEBUG log, to provide that context.

This marks the end of the "training wheels" part of the workshop. Now that
we have a somewhat solid foundation to build on, we can start to make
informed decisions that help us survive the only constant in software
engineering... changing business requirements >_<!

We still have a number of concerns that we can't address within the
the existing API deisgn... but now we have the tools to inform
our server evolution. GEDIUP!



### 21. A note on selling change [[top]](#how-to-inherit-a-mess)

* pkg: **piqued**
* commit: **[github](https://github.com/jsteenb2/mess/commit/67e349e830073a0bcb1adcbe246778e6bcf32a2d)**

```shell
git checkout 67e349e
```

<details>
<summary>Files impacted</summary>



</details>

Your efforts to right the ship needs to be sold to a number of stakeholders.
Each group of stakeholders want to know how they will benefit from your efforts.
Managers are often the biggest decision makers but are typically
the simplest to persuade as well. With management and other non-engineerng IC
stakeholders, the first thing we need to do is to stop using the phrase tech
debt. This phrase is not prioritized with these stakeholders.
Let's first take look through the eyes of management:

1. What do managers care about?
    * They care about getting their teams tasks done

2. What does your manager's manager care about?
    * About all their teams getting their tasks done

3. What does the manager's manager's manager's manager care about?
    * Making their mark... and that requires everyone finishing their tasks

<img style="display: block; margin: auto;" src="piqued/interest_piqued.gif" alt="my interest is piqued">

The common thread when talking to management is that they are mostly bean
counters wanting to make their bosses happy 👻. I'm mostly joking,
but I imagine its not far off for a lot of you. For the rest of you all
that have competent leadership, make it easy for your management to
sell your efforts up the chain of command!

When we're talking to these stakeholders, we need to phrase our endeavor in
terms of improved velocity, improved quality, reduced defects, and an
overall win for every effort going forward. It doesn't have to be hard
and fast numbers to support this.

Take a small example of a task your team has to complete. Write down
what's required to get it done following the status quo/legacy software.

Here are some items to capture:

* What's the amount of work involved?
   * Take a guess, t-shirt size it
* Take note of how much rework/non-value added work is required to get this done
   * Non-value work are often necessary, but mountains of it are eveidence of a systemic problem
* Write down how much risk is inherent to the part of the system impacted by the task
* How easy is it to test?
   * If its difficult or impossible to test we're increasing our risk of defect dramatically
   * If you have data, even if its anectodotal, speak to that
* Do we have a safe means to rollback in the event of an issue?
   * What's the likelihood of encountering that situation?
* Take note of the concern from the team regarding the work
   * Is there a general feeling we're dealing with a house of cards?
* Acknolwedge the on call experience
   * Has it been getting better/worse/stayed the same?

Taking a few moments to answer the above helps you build a coalition around
your new design endeavors to your non-engineering stakeholders. Notice we
never mentioned anything about CPUs/jenkins/kubernetes/etc. We kept it at
the language of management.

Take a moment to explore what this task will look like when you have the
redesign in place. Do not include the time it takes to bootstrap
that new design. We'll get to that *cost* in a bit. Capture the same items
from the list above, only this time do so as if you have your legacy
replacement in place. I recommend making one addition to the list above.
I recommend you add a line item for things that you can now do that
were once _impossible_. The more meaningful that item is to your stakeholder
the better they'll receive the whole of your argument for change.

Quick note about the *cost* of the legacy replacement. If you start by
talking about the cost of the legacy replacement, you're going to be facing an
uphill battle. Management will miss the forest for the trees. Your
management's head will immediately go through the thought exercise of trying
to explain something they don't fully understand to their boss. Not only is
it a design they likely won't fully understand, they'll also have no idea
what true *value* they'll get out of the design. Instead postpone that
discussion until **after** you've provided the *value* of the work.

Now all you have left to do is add everything up and summarize it by stakeholders.
Stakeholders include but are not limited to the following:

<details><summary>Management</summary>

Note how much your velocity could improve across the board. Also,
make note of the ability to test and assure quality earlier in the code's
lifecycle. Ideally, you're able to create a strong integration test footprint.
This raises the bar of quality you're delivering to your stakeholders. If
you take it a step farther and include end to end tests and continuous
deployment tests.... you're in an **insanely strong** position.

<details><summary>On Continuous Deployment Tests</summary>

After I had learned to design and test well I still found issues in
deployment. We aren't deploying in a vaccuum. If you're deploying in
the cloud, you have some sort of cloud infrastructure that has to be
available/setup to do what you need to do. Regardless of where or how
you deploy, you have some environmental dependencies. Any of those can
create an issue when they do not line up with expecations. There's a
lot more to the problem than what we see in a continuous integration (CI)
pipeline.

If you do not have continuous deployment tests, try creating a CLI/scripts
that codifies your deployment tests. These can be issued after a deployment
reducing the risk of defects. Since the tests are codified into a CLI/scripts,
your whole team has a simple way to repeat the process. Once again, we're creating
standardized work. This can take you quite far. When you have the means to deploy
tests that run in your deployment environments, you can take what's in that
CLI/script and reuse it in the tests you setup. A giant standardized win!

</details>

These test feedback loops improve your ability to understand the system.
Increasing the bus factor beyond one is a huge win. That is a (often
undisclosed) risk associated with any team. If a team has high attrition
rate and its largely due to the dysfunction created by the miserable
developer experience and the flawed shortcuts they often produce... you
can help move the needle here by improving the entire team's experience
working within the system. Retention rates are often tied to management
compensation, so if you can make an impact here, you're directly improving
your management's position. The manager has skin in the game to make sure you
can succeed.

When you're talking to management, its about the broader picture. Its
about improving how team as a whole can react to the needs of the business.
Improving your team's feedback loop's in the dev cycle, improves the
bottom line. The more successful the engineers become, the more successful
the manager becomes. Attrition drops, comraderie increases, and you're
seeing a greater impact on the company's performance! Its a win-win situation.
Make sure you sell it as such!

</details>

<details><summary>Engineers</summary>

This is typically an easy sell. If you are feeling the pain, then
undoubedbly others on the team are feeling it as well. When you're
providing a rope to climb out of this mess, they tend to listen. Not
all engineers will entertain the thought of something new. Realistically,
you're going to face engineers wearing at least one of the three hats
mentioned in the [introduction](#1-introduction-top). The folks who have
acclimated, are the biggest resistance you'll face. These often include
managers :sigh:. Often times they understand the tribal knowledge and
believe the struggle they went through to get to this point, are what
everyone should go through. You'll know survival bias when you see it.
The thing to remember with those who have acclimated, is that they
may have some knowledge that could help you improve the bottom line, but
they withold it knowningly or unknowingly. Providing them a space to air
their grievances with a potential new shiny design and the existing
design will prove fruitful more often than not. It'll also help build
trust with those who have concerns about any *changes*. Providing
a tiny prototype to show off the wares your selling, goes a **long**
way with engineers. If they can taste the sweetness of this new
design, they will back it and they will fight for it.

</details>

<details><summary>Yourself</summary>

The first thing you have to remind yourself.... start small! Don't
try and boil the ocean. Build up incrementally wherever possible. This
makes it so much easier to vet assumptions, provide progress to your
other stakeholders, and reaffirm your commitment to the new design. Those
small milestones are motivating. When you identify a bad assumption, its
much easier to rethink your inks when its a small change. As the old adage
goes:

> The only thing youre guaranteed with a big bang design, is a big bang

After you've established the first few updates, you'll hit the ground
running. Remember, its a work in *progress*. You'll be working on progress
for a while if you want to completely sunset the legacy code. If you
don't mind having multiple versions at a time, you can use the 80/20 rule
and pick out the small subset of behavior from the legacy design and replace
it. Once those are in place, and you and the team reap the benefits,
it'll be hard to stop the team from finishing the rest in my experience 🙃.

</details>



### 22. Add v2 Server create API [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/2bb059089efc0d0d8f70b800225fa3c381545d97)**

```shell
git checkout 2bb0590
```

<details>
<summary>Files impacted</summary>

| File                                                 | Count   | Diff                                                                                                                 |
|------------------------------------------------------|---------|----------------------------------------------------------------------------------------------------------------------|
| [allsrv/errors.go](allsrv/errors.go)                 | **9**   | <span style="color:green">++++++</span><span style="color:red">---</span>                                            |
| [allsrv/server.go](allsrv/server.go)                 | **52**  | <span style="color:green">+++++++++++++++++++++++++++++++++++</span><span style="color:red">-----------------</span> |
| [allsrv/server_v2.go](allsrv/server_v2.go)           | **381** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                               |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go) | **128** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                               |


</details>

Here we're leaning into a more structured API. This utilizes versioned URLs
so that our endpoints can evolve in a meaningful way. We also make use of
the [JSON-API spec](https://jsonapi.org/). This provides a common structure
for our consumers. JSON-API is more opinionated than other API specs, but
it has client libs that are widely available across most languages. We've
chosen not to implement the entire spec, but enough to show off the
core benefits of using a spec (not limited to JSON-API spec). The JSON-API
spec is VERY structured (for better or worse), and would make this a
[level 3 RMM](https://www.crummy.com/writing/speaking/2008-QCon/) compliant service,
when the links/relationships are included. That can be incredibly powerful.

As maintainers/developers we get the following from using an API Spec:

  * Standardized API shape, provides for strong abstractions
  * With a spec/standardization you can now remove the boilerplate altogether
    and potentially generate :allthetransportthings: with simple tooling
  * We eliminate some bike-shedding about API design. Kind of like `gofmt`,
    the API is no one's favorite, yet the API is everyone's favorite

Consumers benefit in the following ways:

  * A surprised consumer is an unhappy consumer, following a Spec (even a
    bad one), helps inform consumers and becomes simpler over time to
    reason about.
  * Consumers may not require any SDK/client lib and can traverse the API
    on their own. This is part of the salespitch for RMM lvl 3/JSON-API,
    though I'm not in 100% agreement that is worth the effort.

We've introduced a naive URI versioning scheme. There are a lot of ways
to slice this bread. The simplest is arguably the URI versioning scheme,
which is why we're using it here. However, there are a number of other
options available as well. Versioning is a tough pill to swallow for most
orgs. There are many strategies, and every strategy has 1000x opinions about
why THIS IS THE WAY. Explore the links below yourself, determine what's
important to your organization and go from there.

Take note, there are many conflicting opinions in the resources above :hidethepain:.
Another thing to take note of here is our use of middleware has increased to
include some additional checks. In this case we have some additional checks,
that all return the same response (via the API spec), and creates a one stop
shop for these orthogonal concerns.

For flavor, we've made use of generics to adhere to not only the JSON-API
spec, but also the reduce the boilerplate in dealing with handlers. We'll
expand on this in a bit.

Next we'll take a look at making our tests more flexible so that we can
extend our testcases without having to duplicate the entire test.



<details><summary>References</summary>

3. [Intro to Versioning a Rest API](https://www.freecodecamp.org/news/how-to-version-a-rest-api/)
4. [Versioning Rest Web API Best Practices - MSFT](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design#versioning-a-restful-web-api)
5. [API Design Cheat Sheet](https://github.com/RestCheatSheet/api-cheat-sheet#api-design-cheat-sheet)

</details>

### 23. Refactor v2 tests with table tests [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/99c341dd9e183debe7cbf9b3e1245ae19c15756a)**

```shell
git checkout 99c341d
```

<details>
<summary>Files impacted</summary>

| File                                                 | Count   | Diff                                                                                                                  |
|------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/errors.go](allsrv/errors.go)                 | **6**   | <span style="color:green">++++++</span>                                                                               |
| [allsrv/server_v2.go](allsrv/server_v2.go)           | **144** | <span style="color:green">+++++++++++++++++++++++++++++++</span><span style="color:red">----------------------</span> |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go) | **282** | <span style="color:green">++++++++++++++++++++++++++++++++++++++</span><span style="color:red">---------------</span> |


</details>

There are a few things to note here:

  1. We are making use of table tests. With minimal "relearning" we can
     extend the usecases to accomodate the growing needs of the server.
  2. The test cases make use of some simple helper funcs to make the tests
     more readable. Tests like these act as documentation for future
     contributors. This should not be taken lightly... as the future you,
     will thank you :-).
  3. We make use of a _want func_ here. This might be new to some folks,
     but this is preferred when creating table tests for any UUT that
     returns more than a simple output (i.e. strings/ints/structs). With
     the _want func_, we  get much improved test error stack traces. The entire
     existence of the test is within the usecase. The common test bootstrapping
     is found within the `t.Run` function body, however, it is not a place we
     will find error traces as there is no where for it to fail `:thinker:`.

     We're able to run more assertions/check than what the server responds
     with. For example, checking that the database does not contain a record that should
     not exist.

     However, all that pales in comparison to how much this simplifies the
     logic of the test. You may have run into a situation where table tests
     are paired with a incredibly complex test setup, interleaving multiple
     codepaths for setup and assertions. This is a heavy burden for the next
     person to look at this code. That next person, may be you... look out
     for you.

With the improved test suite, we can make some foundational fixes to align
with JSON-API. The previous implementation did not use a `Data` type for the request body
but with some tests in place, we can validate the desired shape.

Now that we have a solid foundation to build from, we can extend our use case
further to support the read/update/delete APIs. Branch off of this commit
and attempt to add these new APIs.



### 24. Extend ServerV2 API with read/update/delete [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/352965b8e6ccc55549d2ec78a7f571672064b219)**

```shell
git checkout 352965b
```

<details>
<summary>Files impacted</summary>

| File                                                 | Count   | Diff                                                                                                                  |
|------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/db_inmem.go](allsrv/db_inmem.go)             | **6**   | <span style="color:green">++++++</span>                                                                               |
| [allsrv/server.go](allsrv/server.go)                 | **1**   | <span style="color:green">+</span>                                                                                    |
| [allsrv/server_v2.go](allsrv/server_v2.go)           | **178** | <span style="color:green">++++++++++++++++++++++++++++++++++++++</span><span style="color:red">---------------</span> |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go) | **458** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++</span><span style="color:red">------</span> |


</details>

With this little addition, we're at a place where we have a bit of comfort
making changes to the `ServerV2` API. Now that we have a foundation for the
`ServerV2` API, we can finally see how it all comes together in the server
daemon.



### 25. Add simple server daemon for allsrv [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/4cfc573e6c985cc15d5345bbae1e2b9d431dca2e)**

```shell
git checkout 4cfc573
```

<details>
<summary>Files impacted</summary>

| File                                                   | Count  | Diff                                                                  |
|--------------------------------------------------------|--------|-----------------------------------------------------------------------|
| [allsrv/cmd/allsrv/main.go](allsrv/cmd/allsrv/main.go) | **36** | <span style="color:green">++++++++++++++++++++++++++++++++++++</span> |


</details>

To run the daemon run the following from the root of the git repo:

```shell
go run ./allsrv/cmd/allsrv
```

A sample query to get started:

```shell
curl -u admin:pass -H 'Content-Type: application/json' -X POST http://localhost:8091/v1/foos --json '{"data":{"type":"foo","attributes":{"name":"the first","note":"a note"}}}'
```

Play around with the daemon. Hit both v1 and v2 APIs. How do the
two `Server` versions compare?

Now that we have a sample server up and running, let's discuss what it
looks like to start moving on from the original mess.

The first thing we need to do is start communicating our intent to deprecate,
and eventually sunset the original APIs. I can't stress this enough,
**communication is key**. It should be everywhere and EXTREMELY obvious. In
the code, you can add deprecation/sunset headers.

Let's start there, go on and add deprecation headers to ALL the original
endpoints.



### 26. Add deprecation headers to the v1 endpoints [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/899f6364038721b72feb4174cb912748c8c5d1e6)**

```shell
git checkout 899f636
```

<details>
<summary>Files impacted</summary>

| File                                 | Count  | Diff                                                                              |
|--------------------------------------|--------|-----------------------------------------------------------------------------------|
| [allsrv/server.go](allsrv/server.go) | **17** | <span style="color:green">++++++++++++</span><span style="color:red">-----</span> |


</details>

This is simple enough to add and may go largely unnoticed, but if you
control the SDK/client, then you are capable of throwing warning/errors
based on the deprecation headers in the response. If it's a while before
the deprecation deadline, then warn the user, if it's past due... start
throwing some error logs if the endpoints continue to work. There is no
guarantee the endpoint will remain available.

This is a really hard problem. In an ideal world, you own the sdk your
consumers use to interact with your API. This is an excellent place to be
because you can add the glue necessary to transition APIs. Please decorate
your SDKs with useful origin and user agent headers so you're able to target
your deprecation efforts. Here's a look at an example log that gives us
useful `Origin` and `User-Agent` our SDK would add to help us transition
external teams utilizing the SDK:

```json
{
  "level": "INFO",
  "trace_id": "$TRACE_ID",
  "user_id": "$USER_ID",
  "name": "allsrvd.svc",
  "foo_id": "$FOO_ID",
  "origin": "$OTHER_TEAM_SERVICE_NAME",
  "user-agent": "allsrvc (github.com/jsteenb2/allsrvc) / v0.4.0",
  "msg": "foo deleted successfully"
}
```

Knowing who you are serving and what versions of your SDK are being used
is insanely helpful. They only have to approve and merge it once if their
tests pass. With the log above, you can use your logging infrastructure
to visualize what teams are doing with your service and how dated their integrations
might be by utilizing the `User-Agent`. The origin helps you track what services
are hitting you as well `:chef_kiss:`. You can simplify the work for your internal teams
by updating their code to the newest version and cutting them a PR. That's how to
be considerate of your fellow teams while improving your own situation!  We'll go over an
example of how to decorate your SDKs with the information above in a latercommmit.

If you're relying on an OpenAPI spec, just be warned, if your clients have
to generate from that OpenAPI spec... and you have multiple similar endpoints
defined in that OpenAPI spec... your users may be scratching their heads
with which one to use. Some due diligence and a mountain of patience goes
a long way.

Now that we've established our upcoming deprecations, let's take aim at
improving our database story. Say we want to create a SQL DB or some new
integration.

Where do we start? Take a moment to work through the next steps. We'll
then move onto adopting a sqlite db implementation. This should cause
zero changes to the server, and should be 100% dependency injected into
the servers same as the in-mem db is being done now.



### 27. Add db test suite [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/3cde05e889c32d32c6c050d4ac7cc0e7e167d075)**

```shell
git checkout 3cde05e
```

<details>
<summary>Files impacted</summary>

| File                                               | Count   | Diff                                                                                                                  |
|----------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/db_inmem_test.go](allsrv/db_inmem_test.go) | **264** | <span style="color:green">+</span><span style="color:red">----------------------------------------------------</span> |
| [allsrv/db_test.go](allsrv/db_test.go)             | **530** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/server.go](allsrv/server.go)               | **10**  | <span style="color:green">+++++</span><span style="color:red">-----</span>                                            |


</details>

By adding a DB test suite, we've effectively lowered the bar to entry for
any additional DBs we wish to add. With this test suite, we can re-run it
against a SQL DB, nosql db, etc. and expect it to satisfy the behavior the
tests cover. Interestingly, this is not limited to dbs. There are a number
of great examples of test suites in open source software. Here are a couple
examples:

  * [influxdb](https://github.com/influxdata/influxdb/tree/v2.0.0/testing):
              things to note here are the abstraction around KV and the
              service behavior. However, reusing the pkg name, testing,
              is not something i'd advise doing. It will create confusion.
  * [vice](https://github.com/matryer/vice/blob/master/vicetest/test.go):
          this lib does a great job abstracting over the expected _behavior_
          similar to influxdb. When you look through the test for each queue
          you should see the `vicetest` pkg's exported tests being called.
          The pkg name, vicetest, is very explicit, would highly recommend
          using a similar naming strategy.

The key thing here is the language of the database interface. Just as
the two examples above are abstracting over the behavior, we do the same
here. Since the closest thing we have to a domain type is the `Foo` type,
we utilize that as our domain language. Any db will need to be able take
a domain Foo and persist it in using whatever implementation they desire. We aren't
bleeding any details beyond the point of implementation. To illustrate this,
I removed the GORM struct tags, as the domain type should not be limited
by the db design.

Now that we have a little test suite stood up, go on and add a new db implementation
and make sure it passes the test suite. I will be adding a sql store, with
sqlite, as a example to build from, however, feel free to explore this problem
space however you'd like.



<details><summary>References</summary>

6. [InfluxDB Testing](https://github.com/influxdata/influxdb/tree/v2.0.0/testing)
7. [Vice Test Suite](https://github.com/matryer/vice/blob/master/vicetest/test.go)

</details>

### 28. Add sqlite db implementation and test suite [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/8aeed097861fa1ab32f9ffe9e415d30f5705b23c)**

```shell
git checkout 8aeed09
```

<details>
<summary>Files impacted</summary>

| File                                                                                             | Count   | Diff                                                                                                                  |
|--------------------------------------------------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/cmd/allsrv/main.go](allsrv/cmd/allsrv/main.go)                                           | **52**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++</span><span style="color:red">-</span>  |
| [allsrv/db_inmem.go](allsrv/db_inmem.go)                                                         | **2**   | <span style="color:green">+</span><span style="color:red">-</span>                                                    |
| [allsrv/db_sqlite.go](allsrv/db_sqlite.go)                                                       | **121** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/db_sqlite_test.go](allsrv/db_sqlite_test.go)                                             | **50**  | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                   |
| [allsrv/db_test.go](allsrv/db_test.go)                                                           | **115** | <span style="color:green">++++++++++++++++++++++++++++++</span><span style="color:red">-----------------------</span> |
| [allsrv/errors.go](allsrv/errors.go)                                                             | **16**  | <span style="color:green">++++++++++++++</span><span style="color:red">--</span>                                      |
| [allsrv/migrations/migrations.go](allsrv/migrations/migrations.go)                               | **15**  | <span style="color:green">+++++++++++++++</span>                                                                      |
| [allsrv/migrations/sqlite/0001_genesis.down.sql](allsrv/migrations/sqlite/0001_genesis.down.sql) | **1**   | <span style="color:green">+</span>                                                                                    |
| [allsrv/migrations/sqlite/0001_genesis.up.sql](allsrv/migrations/sqlite/0001_genesis.up.sql)     | **8**   | <span style="color:green">++++++++</span>                                                                             |
| [allsrv/server.go](allsrv/server.go)                                                             | **2**   | <span style="color:green">+</span><span style="color:red">-</span>                                                    |
| [go.mod](go.mod)                                                                                 | **9**   | <span style="color:green">+++++++++</span>                                                                            |
| [go.sum](go.sum)                                                                                 | **25**  | <span style="color:green">+++++++++++++++++++++++++</span>                                                            |


</details>

The test suite underwent a few changes here. There are race conditions
in the original test suite, when it comes to executing the concurrency
focused tests. Made a small update here to address that. TL/DR the race is with
the actual `*testing.T` type, so we make use of the closure to capture the
error and log. This enforces the `testing.T` access is done **after** any test
behavior.

We are also able to update the error handling a bit here. I don't care much
for what the error message says, but I care deeply about the behavior of
the errors I receive. I want to validate the correct behavior is obtained.
This is very useful when integrating amongst a larger, more complex system.
For this trivial `Foo` server, we don't have much complexity.

The sqlite db implementation here is fairly trivial once again. We're able
to reuse the test suite in full. All that was required is a new funciton to initize the unit under test (UUT),
and the rest is including 4 more lines of code to call the test around it.
Not to shabby.

Since we've effectively decoupled our domain `Foo` from the db entity `foo`,
we've provided maximum flexibility to our database implementation without
having to pollute the domain space. This is intensely useful as a system
grows.

Think through, what would it look like to add a `PostgreSQL` db implementation?
Not as much now that you have a test suite to verify the desired behavior.

The last thing that is missing here is what we do to decouple our server
from HTTP. There is a glaring hole in our design, and that's the lack of
service layer. The layer where all our business logic resides. Take a moment
to think through what this might look like.

#### Questions

* [ ] How would you break up the server so that it's no longer coupled to HTTP/REST?
* [ ] What can we do to allow ourselves to support any myriad of `RPC` technologies without duplicating allthe business logic?


### 29. Add the service layer to consolidate all domain logic [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/1fb7dbfe39181a33c2f015180b90e53facb9ce6f)**

```shell
git checkout 1fb7dbf
```

<details>
<summary>Files impacted</summary>

| File                                                   | Count   | Diff                                                                                                                  |
|--------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/cmd/allsrv/main.go](allsrv/cmd/allsrv/main.go) | **55**  | <span style="color:green">++++++++++++++++++++++++++++++++++</span><span style="color:red">-------------------</span> |
| [allsrv/errors.go](allsrv/errors.go)                   | **6**   | <span style="color:green">++++++</span>                                                                               |
| [allsrv/observe_db.go](allsrv/observe_db.go)           | **8**   | <span style="color:green">++++</span><span style="color:red">----</span>                                              |
| [allsrv/server.go](allsrv/server.go)                   | **61**  | <span style="color:green">++++++++++++++++++++++++++</span><span style="color:red">---------------------------</span> |
| [allsrv/server_test.go](allsrv/server_test.go)         | **16**  | <span style="color:green">++++++++</span><span style="color:red">--------</span>                                      |
| [allsrv/server_v2.go](allsrv/server_v2.go)             | **79**  | <span style="color:green">+++++++++++++++++++++</span><span style="color:red">--------------------------------</span> |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go)   | **32**  | <span style="color:green">+++++++++++++++++</span><span style="color:red">---------------</span>                      |
| [allsrv/svc.go](allsrv/svc.go)                         | **118** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/svc_mw_logging.go](allsrv/svc_mw_logging.go)   | **96**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/svc_observer.go](allsrv/svc_observer.go)       | **72**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |


</details>

This might seem like a "moving the cheese" change. However, upon closer look
we see that the `server_v2` implementation is purely a translation between
the HTTP RESTful API and the domain. All traffic speaks to the service,
which holds all the logic for the Foo domain.

We've effectively decoupled the domain from the transport layer (HTTP).
Any additional transport we want to support (gRPC/Thrift/etc) is merely
creating the transport implementation. We won't duplicate our logic in each transport layer.
Often, when we have consolidated all the business logic, it's very simple to just generate the RPC layer and inject
the SVC to transact with the different API integrations.



### 30. Provide test suite for SVC behavior and fill gaps in behavior [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/41b1cd4dade78b9e0153b9917048902e37b29748)**

```shell
git checkout 41b1cd4
```

<details>
<summary>Files impacted</summary>

| File                                                 | Count   | Diff                                                                                                    |
|------------------------------------------------------|---------|---------------------------------------------------------------------------------------------------------|
| [allsrv/errors.go](allsrv/errors.go)                 | **33**  | <span style="color:green">+++++++++++++++++++++++++++</span><span style="color:red">------</span>       |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go) | **6**   | <span style="color:green">+</span><span style="color:red">-----</span>                                  |
| [allsrv/svc.go](allsrv/svc.go)                       | **39**  | <span style="color:green">+++++++++++++++++++++++++</span><span style="color:red">--------------</span> |
| [allsrv/svc_mw_logging.go](allsrv/svc_mw_logging.go) | **10**  | <span style="color:green">++++++</span><span style="color:red">----</span>                              |
| [allsrv/svc_suite_test.go](allsrv/svc_suite_test.go) | **463** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                  |
| [allsrv/svc_test.go](allsrv/svc_test.go)             | **22**  | <span style="color:green">++++++++++++++++++++++</span>                                                 |


</details>

This test suite provides the rest of the codebase super powers. When we have
new requirements to add, we can extend the service's test suite and we can use
that implementation across any number of implementation details. However,
the super power shows up when we start to integrate multiple service
implementations.

Here's a thought exercise:

Perhaps there is a `Bar` service that integrates the `Foo` service as part of a modular
monolith design. Now you have scale/requirements hitting you that force you
to scale `Foo` independent of `Bar` or vice versa. Perhaps we pull out the foo
svc into its own deployment. Now our `Bar` service needs to access the `Foo` service
via some RPC channel (HTTP|REST/gRPC/etc.).

We then create a remote SVC implementation, perhaps an HTTP client, that
implements the SVC behavior (interface). How do we verify this adheres to
the service behavior? Simple enough, just create another test with the
service's test suite, initilaize the necessary components, and excute them...
pretty simple and guaranteed to be in line with the previous implementation.

#### Questions

* [ ] Now that we have a remote SVC implementation via an HTTP client, what else
might we want to provide?
* [ ] How about a CLI that can integrate with the `Foo`?
* [ ] How would we test a CLI to validate it satisfies the `SVC` interface?


### 31. Add http client and fixup missing pieces [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/808eb0fd7b577236d98e1f19ec23a59186183c9b)**

```shell
git checkout 808eb0f
```

<details>
<summary>Files impacted</summary>

| File                                                 | Count   | Diff                                                                                   |
|------------------------------------------------------|---------|----------------------------------------------------------------------------------------|
| [allsrv/client_http.go](allsrv/client_http.go)       | **192** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [allsrv/errors.go](allsrv/errors.go)                 | **12**  | <span style="color:green">++++++++++++</span>                                          |
| [allsrv/server_v2.go](allsrv/server_v2.go)           | **2**   | <span style="color:green">+</span><span style="color:red">-</span>                     |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go) | **12**  | <span style="color:green">++++++++++++</span>                                          |
| [allsrv/svc.go](allsrv/svc.go)                       | **6**   | <span style="color:green">++++++</span>                                                |
| [allsrv/svc_suite_test.go](allsrv/svc_suite_test.go) | **22**  | <span style="color:green">+++++++++++++++++++++</span><span style="color:red">-</span> |
| [allsrv/svc_test.go](allsrv/svc_test.go)             | **20**  | <span style="color:green">++++++++++++</span><span style="color:red">--------</span>   |


</details>

Here we've added the HTTP client. Again, we're pulling from the standard library
because it's a trivial example. Even with this, we're able to put together a client
that speaks the languae of our domain, and fulfills the behavior of our SVC.
We've provided a fair degree of confidence by utilizing the same `SVC` test
suite we had with the `Service` implementation itself. To top it all off,
we're able to refactor our tests a bit to reuse the constructor for the
`SVC` dependency, leaving us with a standardized setup.

With standardized tests you benefit of reusing the tests. Additionally,
any new contributor only needs to understand a single test setup,
and then writes a testcase. Its extremely straightforward after the
initial onboarding.

Last commit message, I spoke of adding a CLI companion to this. With the HTTP
client we've just created, go on and create a CLI and put it under test
with the same test suite :yaaaaaaaaaas:!



### 32. Add allsrvc CLI companion [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/32a17c0e5c87fac5459ce7fceea259197d27ec00)**

```shell
git checkout 32a17c0
```

<details>
<summary>Files impacted</summary>

| File                                                                                                 | Count   | Diff                                                                                                                  |
|------------------------------------------------------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/allsrvtesting/service_inmem.go](allsrv/allsrvtesting/service_inmem.go)                       | **34**  | <span style="color:green">++++++++++++++++++++++++++++++++++</span>                                                   |
| [allsrv/allsrvtesting/utils.go](allsrv/allsrvtesting/utils.go)                                       | **26**  | <span style="color:green">++++++++++++++++++++++++++</span>                                                           |
| [allsrv/cmd/allsrvc/main.go](allsrv/cmd/allsrvc/main.go)                                             | **147** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/cmd/allsrvc/main_test.go](allsrv/cmd/allsrvc/main_test.go)                                   | **79**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span>                                |
| [allsrv/db_test.go](allsrv/db_test.go)                                                               | **11**  | <span style="color:green">++++++</span><span style="color:red">-----</span>                                           |
| [allsrv/errors.go](allsrv/errors.go)                                                                 | **2**   | <span style="color:green">+</span><span style="color:red">-</span>                                                    |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go)                                                 | **77**  | <span style="color:green">++++++++++++++++</span><span style="color:red">-------------------------------------</span> |
| [allsrv/svc_suite_test.go => allsrv/allsrvtesting/test_suite.go](allsrv/allsrvtesting/test_suite.go) | **125** | <span style="color:green">++++++++++++++++++++++++++</span><span style="color:red">---------------------------</span> |
| [allsrv/svc_test.go](allsrv/svc_test.go)                                                             | **19**  | <span style="color:green">+++</span><span style="color:red">----------------</span>                                   |
| [go.mod](go.mod)                                                                                     | **3**   | <span style="color:green">+++</span>                                                                                  |
| [go.sum](go.sum)                                                                                     | **8**   | <span style="color:green">++++++++</span>                                                                             |


</details>

Again here we start with a little refactoring. This time we're creating
a new `allsrvtesting` pkg to hold the reusable code bits. Now we can
implement our CLI and then create the svc implemenation utilizing the cli
in our tests. With a handful of lines of code we're able to create a
high degree of certainty in the implementation of our CLI. The
implementation and testing both, are not limited by the SVC behavior.
You can extend the CLI with additional behavior beyond that of the SVC.
Additional tests can be added to accomodate any additional behavior.

Here we are at the end of the session where we've matured an intensely
immature service implementation. We've covered a lot of ground. We can
sum it up quickly with:

  1. Tests provide certainty we've not broken the existing implementation
  2. Versioning an API helps us transition into the replacement
      * note: we determined in that the v1 was not serving our users
              best interest, so we moved onto a structured JSON-API spec
  3. Observability is SUUUPER important. Just like with testing, we want
     to keep a close eye on our metrics. We want to make sure our changes
     are improving the bottom line. Without this information, we're flying
     blind.
  4. Isolating the SVC/usecases from the transport & db layers gives us
     freedom to reuse that business logic across any number of transport
     & db layers, improve our observability stack along the way, and allow
     us to create a reusable test suite that is usable across any
     implementation of the SVC. With that test suite, creating and
     verifying a new SVC implementation.



### 33. Add pprof routes [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/ddafbe2ba3566d43a35d302e75bd544364edfc4b)**

```shell
git checkout ddafbe2
```

<details>
<summary>Files impacted</summary>

| File                                                   | Count | Diff                                      |
|--------------------------------------------------------|-------|-------------------------------------------|
| [allsrv/cmd/allsrv/main.go](allsrv/cmd/allsrv/main.go) | **8** | <span style="color:green">++++++++</span> |


</details>

One last bit of important data we're missing is our profiling. This can be
applied at any time in the development. Its added here last because there is
already a lot to cover before thi. However, with these endpoints, you can
create profiles that will breakdown difference performance characteristics
of your system. One of the beautiful things here is you can grab the
profiles ad-hoc or create a recurring drop to grab profiles at different
intervals. Regardless when or how you do it, it all goes through the
HTTP API :chef_kiss:.

See the references below for a a number of good resources for [pprof](https://pkg.go.dev/runtime/pprof).



<details><summary>References</summary>

8. [pprof tool package](https://pkg.go.dev/runtime/pprof)
9. [pprof HTTP integration](https://pkg.go.dev/net/http/pprof)
10. [Profiling Go Programs - Russ Cox](https://go.dev/blog/pprof)

</details>

### 34. Add github.com/jsteenb2/errors module to improve error handling [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/baf5cd9b4cb6bcfd978a8b1fa12cc96222df5fa5)**

```shell
git checkout baf5cd9
```

<details>
<summary>Files impacted</summary>

| File                                                                           | Count   | Diff                                                                                                                  |
|--------------------------------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/allsrvtesting/service_inmem.go](allsrv/allsrvtesting/service_inmem.go) | **6**   | <span style="color:green">++++</span><span style="color:red">--</span>                                                |
| [allsrv/allsrvtesting/test_suite.go](allsrv/allsrvtesting/test_suite.go)       | **19**  | <span style="color:green">++++++++++</span><span style="color:red">---------</span>                                   |
| [allsrv/client_http.go](allsrv/client_http.go)                                 | **23**  | <span style="color:green">++++++++++++++++</span><span style="color:red">-------</span>                               |
| [allsrv/db_sqlite.go](allsrv/db_sqlite.go)                                     | **31**  | <span style="color:green">+++++++++++++++++++++</span><span style="color:red">----------</span>                       |
| [allsrv/db_sqlite_test.go](allsrv/db_sqlite_test.go)                           | **14**  | <span style="color:green">++++++++</span><span style="color:red">------</span>                                        |
| [allsrv/db_test.go](allsrv/db_test.go)                                         | **10**  | <span style="color:green">+++++</span><span style="color:red">-----</span>                                            |
| [allsrv/errors.go](allsrv/errors.go)                                           | **104** | <span style="color:green">++++++++++++++++++</span><span style="color:red">-----------------------------------</span> |
| [allsrv/server_v2.go](allsrv/server_v2.go)                                     | **36**  | <span style="color:green">++++++++++++++++++</span><span style="color:red">------------------</span>                  |
| [allsrv/svc.go](allsrv/svc.go)                                                 | **16**  | <span style="color:green">+++++++++</span><span style="color:red">-------</span>                                      |
| [allsrv/svc_mw_logging.go](allsrv/svc_mw_logging.go)                           | **4**   | <span style="color:green">+++</span><span style="color:red">-</span>                                                  |
| [allsrv/svc_test.go](allsrv/svc_test.go)                                       | **7**   | <span style="color:green">+++++++</span>                                                                              |
| [go.mod](go.mod)                                                               | **1**   | <span style="color:green">+</span>                                                                                    |
| [go.sum](go.sum)                                                               | **2**   | <span style="color:green">++</span>                                                                                   |


</details>

The addition of this module gives us radically improved error handling and
logging. We now have the ability to tie into the std lib errors.Is/As
functionality instead of writing it ourselves. Our `ErrKinds` are now
useful for an entire domain.

We've added a touch more structure to our error handling with the new
module. With this new structure we can improve our logging once again.
The better your error handling is, the better your logging will get.

Here's an example pulled from the service create foo with exists test:

```json
{
  "time": "2024-07-05T22:56:16.976262-05:00",
  "level": "ERROR",
  "msg": "failed to create foo",
  "input_name": "existing-foo",
  "input_note": "new note",
  "took_ms": "0s",
  "err": "foo exists",
  "err_fields": {
    "sqlite_err_code": "constraint failed",
    "sqlite_err_extended_code": "constraint failed",
    "sqlite_system_errno": "errno 0",
    "err_kind": "exists",
    "stack_trace": [
      "github.com/jsteenb2/mess/allsrv/svc.go:97[(*Service).CreateFoo]",
      "github.com/jsteenb2/mess/allsrv/db_sqlite.go:38[(*sqlDB).CreateFoo]",
      "github.com/jsteenb2/mess/allsrv/db_sqlite.go:96[(*sqlDB).exec]"
    ]
  }
}
```

As your system gets more and more complex, your errors are capable of
extending not support additional details. The extensibility is amazing
for a growing codebase.



<details><summary>References</summary>

11. [errors module](https://github.com/jsteenb2/errors)

</details>

### 35. Add github.com/jsteenb2/allsrvc SDK module [[top]](#how-to-inherit-a-mess)

* pkg: **allsrv**
* commit: **[github](https://github.com/jsteenb2/mess/commit/9d41492653a5267c453656d04a1f2925bc860737)**

```shell
git checkout 9d41492
```

<details>
<summary>Files impacted</summary>

| File                                                               | Count   | Diff                                                                                                                  |
|--------------------------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| [allsrv/client_http.go](allsrv/client_http.go)                     | **185** | <span style="color:green">+++++++++++++++++</span><span style="color:red">------------------------------------</span> |
| [allsrv/cmd/allsrvc/main.go](allsrv/cmd/allsrvc/main.go)           | **45**  | <span style="color:green">+++++++++++++++++++++++++++++++++++</span><span style="color:red">----------</span>         |
| [allsrv/cmd/allsrvc/main_test.go](allsrv/cmd/allsrvc/main_test.go) | **5**   | <span style="color:green">+++</span><span style="color:red">--</span>                                                 |
| [allsrv/server_v2.go](allsrv/server_v2.go)                         | **277** | <span style="color:green">++++++++++++++++++++</span><span style="color:red">---------------------------------</span> |
| [allsrv/server_v2_test.go](allsrv/server_v2_test.go)               | **184** | <span style="color:green">+++++++++++++++++++++++++++</span><span style="color:red">--------------------------</span> |
| [allsrv/svc_mw_logging.go](allsrv/svc_mw_logging.go)               | **35**  | <span style="color:green">++++++++++++++++++++</span><span style="color:red">---------------</span>                   |
| [go.mod](go.mod)                                                   | **3**   | <span style="color:green">++</span><span style="color:red">-</span>                                                   |
| [go.sum](go.sum)                                                   | **6**   | <span style="color:green">++++</span><span style="color:red">--</span>                                                |


</details>

We fixup our http client to make use of the
[github.com/jsteenb2/allsrvc](https://github.com/jsteenb2/allsrvc)
SDK. As you, we can clean up a good bit of duplication by utilizing
the SDK as a source of truth for the API types. We've broken up the SDK
from the service/server module. Effectively breaking one of the thorniest
problems large organizations with a large go ecosystem face.

When we leave the SDK inside the service module, its forces all the
depdencies of the service onto any SDK consumer. This creates a series
of problems.

1. The SDK creates a ton of bloat in the user's module.
2. The SDK undergoes a lot of version changes when coupled to the service module version.
3. Circular module dependencies are real, and can cause a LOT of pain.
    * Check out [perseus](https://github.com/CrowdStrike/perseus) to help visualize this!
4. If you do it this way, then other teams will also do it this way, putting tremendous pressure on your CI/build pipelines.

Instead of exporting an SDK from your service, opt for a separate module
for the SDK. This radically changes the game. You can use the SDK module
in the `Service` module to remain *DRY*. However, **DO NOT** import the
`Service` module into the SDK module!

Now that we have the tiny SDK module, we're able to obtain some important
data to help us track who is hitting our API. We now get access to the `Origin`
and `User-Agent` of the callee. Here is an example of a log that
adds the [version of the module](https://github.com/jsteenb2/allsrvc/blob/main/client.go#L21-L30)
as part of `User-Agent` and `Origin` headers when communicating with the server:

```json
{
  "time": "2024-07-06T20:46:58.614226-05:00",
  "level": "ERROR",
  "source": {
    "function": "github.com/jsteenb2/mess/allsrv.(*svcMWLogger).CreateFoo",
    "file": "github.com/jsteenb2/mess/allsrv/svc_mw_logging.go",
    "line": 32
  },
  "msg": "failed to create foo",
  "input_name": "name1",
  "input_note": "note1",
  "took_ms": "0s",
  "origin": "allsrvc",
  "user_agent": "allsrvc (github.com/jsteenb2/allsrvc) / v0.4.0",
  "trace_id": "b9106e52-907b-4bc4-af91-6596e98d3795",
  "err": "foo name1 exists",
  "err_fields": {
    "name": "name1",
    "existing_foo_id": "3a826632-ec30-4852-b4a6-e4a4497ddda8",
    "err_kind": "exists",
    "stack_trace": [
      "github.com/jsteenb2/mess/allsrv/svc.go:97[(*Service).CreateFoo]",
      "github.com/jsteenb2/mess/allsrv/db_inmem.go:20[(*InmemDB).CreateFoo]"
    ]
  }
}
```

With this information, we're in a good position to make proactive changes
to remove our own blockers. Excellent stuff!

Additionally, we've imported our SDK into the `Service` module to *DRY* up
the HTTP API contract. No need to duplicate these things as the server is
dependent on the http client's JSON API types. This is awesome, as we're
still able to keep things DRY, without all the downside of the SDK depending
on the Service (i.e. dependency bloat).

Lastly, we update the CLI to include basic auth. Try exercising the new updates.
Use the CLI to issue some CRUD commands against the server. Start the server
first with:

```shell
go run ./allsrv/cmd/allsrv | jq
```

Then you can install the CLI and make sure to add `$GOBIN` to your `$PATH`:

```shell
go install ./allsrv/cmd/allsrvc
```

Now issue a request to create a foo:

```shell
allsrvc create --name first --note "some note"
```

Now issue another create a foo with the same name:

```shell
allsrvc create --name first --note "some other note"
```

The previous command should fail. Check out the output from the `allsrvc`
CLI as well as the logs from the server. Enjoy those beautiful logs!

This marks the end of our time with the `allsrv` package!



<details><summary>References</summary>

12. [SDK module - github.com/jsteenb2/allsrvc](https://github.com/jsteenb2/allsrvc)
13. [Setting version in SDK via debug.BuildInfo](https://github.com/jsteenb2/allsrvc/blob/main/client.go#L21-L30)
14. [Perseus module tracker](https://github.com/CrowdStrike/perseus)

</details>

### 36. Add slimmed down wild-workouts-go-ddd-example repo [[top]](#how-to-inherit-a-mess)

* pkg: **wild-workouts**
* commit: **[github](https://github.com/jsteenb2/mess/commit/ff0cb5af8fea7b2bcfc0700851e5119a71e88e4a)**

```shell
git checkout ff0cb5a
```

<details>
<summary>Files impacted</summary>

| File                                                                                                                                         | Count    | Diff                                                                                   |
|----------------------------------------------------------------------------------------------------------------------------------------------|----------|----------------------------------------------------------------------------------------|
| [wild-workouts/.gitignore](wild-workouts/.gitignore)                                                                                         | **36**   | <span style="color:green">++++++++++++++++++++++++++++++++++++</span>                  |
| [wild-workouts/LICENSE](wild-workouts/LICENSE)                                                                                               | **21**   | <span style="color:green">+++++++++++++++++++++</span>                                 |
| [wild-workouts/Makefile](wild-workouts/Makefile)                                                                                             | **48**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++</span>      |
| [wild-workouts/README.md](wild-workouts/README.md)                                                                                           | **107**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/doc.go](wild-workouts/doc.go)                                                                                                 | **10**   | <span style="color:green">++++++++++</span>                                            |
| [wild-workouts/internal/common/auth/http.go](wild-workouts/internal/common/auth/http.go)                                                     | **84**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/auth/http_mock.go](wild-workouts/internal/common/auth/http_mock.go)                                           | **44**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++</span>          |
| [wild-workouts/internal/common/client/auth.go](wild-workouts/internal/common/client/auth.go)                                                 | **41**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++</span>             |
| [wild-workouts/internal/common/client/grpc.go](wild-workouts/internal/common/client/grpc.go)                                                 | **81**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/net.go](wild-workouts/internal/common/client/net.go)                                                   | **37**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++</span>                 |
| [wild-workouts/internal/common/client/trainer/openapi_client_gen.go](wild-workouts/internal/common/client/trainer/openapi_client_gen.go)     | **568**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/trainer/openapi_types.gen.go](wild-workouts/internal/common/client/trainer/openapi_types.gen.go)       | **57**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/trainings/openapi_client_gen.go](wild-workouts/internal/common/client/trainings/openapi_client_gen.go) | **1004** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/trainings/openapi_types.gen.go](wild-workouts/internal/common/client/trainings/openapi_types.gen.go)   | **60**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/users/openapi_client_gen.go](wild-workouts/internal/common/client/users/openapi_client_gen.go)         | **234**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/client/users/openapi_types.gen.go](wild-workouts/internal/common/client/users/openapi_types.gen.go)           | **15**   | <span style="color:green">+++++++++++++++</span>                                       |
| [wild-workouts/internal/common/decorator/command.go](wild-workouts/internal/common/decorator/command.go)                                     | **27**   | <span style="color:green">+++++++++++++++++++++++++++</span>                           |
| [wild-workouts/internal/common/decorator/logging.go](wild-workouts/internal/common/decorator/logging.go)                                     | **56**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/decorator/metrics.go](wild-workouts/internal/common/decorator/metrics.go)                                     | **62**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/decorator/query.go](wild-workouts/internal/common/decorator/query.go)                                         | **21**   | <span style="color:green">+++++++++++++++++++++</span>                                 |
| [wild-workouts/internal/common/errors/errors.go](wild-workouts/internal/common/errors/errors.go)                                             | **53**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/genproto/trainer/trainer.pb.go](wild-workouts/internal/common/genproto/trainer/trainer.pb.go)                 | **315**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/genproto/trainer/trainer_grpc.pb.go](wild-workouts/internal/common/genproto/trainer/trainer_grpc.pb.go)       | **209**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/genproto/users/users.pb.go](wild-workouts/internal/common/genproto/users/users.pb.go)                         | **305**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/genproto/users/users_grpc.pb.go](wild-workouts/internal/common/genproto/users/users_grpc.pb.go)               | **137**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/go.mod](wild-workouts/internal/common/go.mod)                                                                 | **54**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/go.sum](wild-workouts/internal/common/go.sum)                                                                 | **627**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/logs/cqrs.go](wild-workouts/internal/common/logs/cqrs.go)                                                     | **15**   | <span style="color:green">+++++++++++++++</span>                                       |
| [wild-workouts/internal/common/logs/http.go](wild-workouts/internal/common/logs/http.go)                                                     | **64**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/logs/logrus.go](wild-workouts/internal/common/logs/logrus.go)                                                 | **31**   | <span style="color:green">+++++++++++++++++++++++++++++++</span>                       |
| [wild-workouts/internal/common/metrics/dummy.go](wild-workouts/internal/common/metrics/dummy.go)                                             | **7**    | <span style="color:green">+++++++</span>                                               |
| [wild-workouts/internal/common/server/grpc.go](wild-workouts/internal/common/server/grpc.go)                                                 | **54**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/server/http.go](wild-workouts/internal/common/server/http.go)                                                 | **96**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/server/httperr/http_error.go](wild-workouts/internal/common/server/httperr/http_error.go)                     | **57**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/tests/clients.go](wild-workouts/internal/common/tests/clients.go)                                             | **174**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/tests/e2e_test.go](wild-workouts/internal/common/tests/e2e_test.go)                                           | **69**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/common/tests/hours.go](wild-workouts/internal/common/tests/hours.go)                                                 | **15**   | <span style="color:green">+++++++++++++++</span>                                       |
| [wild-workouts/internal/common/tests/jwt.go](wild-workouts/internal/common/tests/jwt.go)                                                     | **35**   | <span style="color:green">+++++++++++++++++++++++++++++++++++</span>                   |
| [wild-workouts/internal/common/tests/wait.go](wild-workouts/internal/common/tests/wait.go)                                                   | **33**   | <span style="color:green">+++++++++++++++++++++++++++++++++</span>                     |
| [wild-workouts/internal/trainer/adapters/hour_firestore_repository.go](wild-workouts/internal/trainer/adapters/hour_firestore_repository.go) | **222**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/adapters/hour_memory_repository.go](wild-workouts/internal/trainer/adapters/hour_memory_repository.go)       | **69**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/adapters/hour_mysql_repository.go](wild-workouts/internal/trainer/adapters/hour_mysql_repository.go)         | **198**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/adapters/hour_repository_test.go](wild-workouts/internal/trainer/adapters/hour_repository_test.go)           | **350**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/app/app.go](wild-workouts/internal/trainer/app/app.go)                                                       | **23**   | <span style="color:green">+++++++++++++++++++++++</span>                               |
| [wild-workouts/internal/trainer/app/command/cancel_training.go](wild-workouts/internal/trainer/app/command/cancel_training.go)               | **50**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++</span>    |
| [wild-workouts/internal/trainer/app/command/make_hours_available.go](wild-workouts/internal/trainer/app/command/make_hours_available.go)     | **52**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++++</span>  |
| [wild-workouts/internal/trainer/app/command/make_hours_unavailable.go](wild-workouts/internal/trainer/app/command/make_hours_unavailable.go) | **52**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++++</span>  |
| [wild-workouts/internal/trainer/app/command/schedule_training.go](wild-workouts/internal/trainer/app/command/schedule_training.go)           | **50**   | <span style="color:green">++++++++++++++++++++++++++++++++++++++++++++++++++</span>    |
| [wild-workouts/internal/trainer/app/query/hour_availability.go](wild-workouts/internal/trainer/app/query/hour_availability.go)               | **60**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/app/query/types.go](wild-workouts/internal/trainer/app/query/types.go)                                       | **2**    | <span style="color:green">++</span>                                                    |
| [wild-workouts/internal/trainer/domain/hour/availability.go](wild-workouts/internal/trainer/domain/hour/availability.go)                     | **97**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/domain/hour/availability_test.go](wild-workouts/internal/trainer/domain/hour/availability_test.go)           | **125**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/domain/hour/hour.go](wild-workouts/internal/trainer/domain/hour/hour.go)                                     | **221**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/domain/hour/hour_test.go](wild-workouts/internal/trainer/domain/hour/hour_test.go)                           | **248**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/domain/hour/repository.go](wild-workouts/internal/trainer/domain/hour/repository.go)                         | **15**   | <span style="color:green">+++++++++++++++</span>                                       |
| [wild-workouts/internal/trainer/fixtures.go](wild-workouts/internal/trainer/fixtures.go)                                                     | **99**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/go.mod](wild-workouts/internal/trainer/go.mod)                                                               | **62**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/go.sum](wild-workouts/internal/trainer/go.sum)                                                               | **628**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/main.go](wild-workouts/internal/trainer/main.go)                                                             | **45**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++</span>         |
| [wild-workouts/internal/trainer/ports/grpc.go](wild-workouts/internal/trainer/ports/grpc.go)                                                 | **68**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/ports/http.go](wild-workouts/internal/trainer/ports/http.go)                                                 | **49**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++</span>     |
| [wild-workouts/internal/trainer/ports/openapi_api.gen.go](wild-workouts/internal/trainer/ports/openapi_api.gen.go)                           | **161**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/ports/openapi_types.gen.go](wild-workouts/internal/trainer/ports/openapi_types.gen.go)                       | **57**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/internal/trainer/service/application.go](wild-workouts/internal/trainer/service/application.go)                               | **51**   | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++</span>   |
| [wild-workouts/internal/trainer/service/component_test.go](wild-workouts/internal/trainer/service/component_test.go)                         | **108**  | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++++++</span> |
| [wild-workouts/sql/schema.sql](wild-workouts/sql/schema.sql)                                                                                 | **6**    | <span style="color:green">++++++</span>                                                |


</details>

This repo originates from `threedotlabs` to accompany their "Introducing
Clean Architecture" blog post found [here](https://threedots.tech/post/introducing-clean-architecture/).

Take a few moments... gather your thoughts about this slimmed down example.

#### Questions

* [ ] What positivie aspects do you find in this design pattern?
* [ ] What concerns do you have with this design pattern?


<details><summary>References</summary>

15. [Threedots Tech Introduction to Clean Architecture](https://threedots.tech/post/introducing-clean-architecture/)

</details>

### 37. Add thoughts on wild-workouts Clean architecture [[top]](#how-to-inherit-a-mess)

* pkg: **wild-workouts**
* commit: **[github](https://github.com/jsteenb2/mess/commit/6c6c6b34ffe5b60e062b638f71d084d817c4d161)**

```shell
git checkout 6c6c6b3
```

<details>
<summary>Files impacted</summary>

| File                                         | Count  | Diff                                                                                                                  |
|----------------------------------------------|--------|-----------------------------------------------------------------------------------------------------------------------|
| [wild-workouts/doc.go](wild-workouts/doc.go) | **79** | <span style="color:green">+++++++++++++++++++++++++++++++++++++++++++++++++</span><span style="color:red">----</span> |


</details>

Shared with the in-person *workshop*: The tale of **Abu Fawwaz**, a
reminder that the best solution is often the simplest. However, often
times the simple solution is often hidden in plain site.

<details><summary>Spoiler</summary>

<h4>Praises</h4>

1. "Clean" code, good separation of concerns, i.e. no http code is driving the app or repo
   implementation.
    * I like "Clean" code. But I like to fit the design patterns to the language at hand as well. A
      good chunk of "Clean" is necessary b/c of how overly complex class oriented languages make
      modeling your problem space. "Clean" is an attempt at wrangling this kraken so that a team can
      maintain and extend it further. If you can take one thing away from "Clean"/"Hexagonal"/"Onion"
      it's this, empower your design with encapsulation/separation of concerns.
2. gives the system/codebase structure
    * Without an intended design for your system/codebase's structure, you'll end up with a giant
      mess on your hands. We'll explore this in our next example.
    * familiar to folks coming from another language to go
        * b/c it isn't idiomatic go, and looks and feels like Java, it can be more comfortable
	  for folks coming to go from various OO languages that stress "Clean" or similar.
    * provides orthogonal concerns via middleware
	* though, one word of caution on logging middlewares. If _everything_ gets a logging middleware
          it'll make your logging infrastructure VERY noisy. I don't find a ton of value in logging at
	  all layers of a service. I find it more burdensome than helpful most the time. Having errors
	  that provide context, can improve your logging experience DRAMATICALLY.
3. tests at the feature level, this is fantastic.
4. errors are structured
     * can't say enough good things about structured error handling. Adds a ton of value to
       any and all service codebases

<h4>Concerns</h4>

1. whiplash, do we need a ton of pkgs to make code "Clean"?
	* spoiler alert, you do not :-)
2. TOO much "structure"
	* hot take: this is the reason many go devs snear at "clean" go code, b/c its actually
	  messy in practice. With Go's pkg oriented design we don't need to separate by "what is"
	  but rather... what it's trying to accomplish/do. This "Clean" style does wonders for
	  getting consultancies paid... but leads to the inevitable mess with any mixed seniority
	  group.
	  DC did a good writeup on pkg naming [here](https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common)
	* take a count, how many layers of pkgs do you have to jump through to understand how
		  something works?
	* how will a younger more green engineer fair?
3. codebase requires an uphill climb to understand how to work with the "Clean" principles
	* Newer engineers will have an uphill climb to learn the ways of your companies "Clean" implementation.
	  For senior engineers... they'll have worked with "Clean" in one way or another. The eventual
	  bikeshed moments will transpire allowing seniors to debate at ends about meaningless things.
	  These are the two extremes I've run into with "Clean", and I am willing to wager I'm not the
 	  only one who has experienced this.
4. a great example of writing Java code with Go
	* for experienced go devs, this can be extremely draining. The other big downside is...
	  these patterns will "partially" reproduce as the company grows. I say partially,
	  b/c the patterns will become copy pastaed all over with each team adding their own unique
	  idiosyncratic takes.
5. modules for :allthethings:, when you have the following line in your module, you are likely "holding it wrong":

```
replace github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common => ../common/
```

6. The boundaries of modules in this repo don't make sense to me (if you understand it,
   please explain it to me). This one really gives me chills as the entire module within
   each application (trainer/training/common) are exported.... meaning hyrum's law is
   just around the corner... :cringe:
7. common module
	* this harkens back to the issue before where we're using principles that work in other
	  languages, but languish in go. For example, take the decorators pkg... why do we need
	  `decorator.ApplyCommandDecorators` in its own pkg???
	  Contrast the indirection with our first example, "allsrv", where everything is fairly close
	  in proximity and implementation. Its radically simpler to juggle in an average developer
	  like me's head. Don't require a 200 IQ to contribute.
	* Why do we need a ApplyCommandDecorators in the common/metrics pkg instead of keeping
	  this encapsulated inside the app/command pkg?
8. tests are missing for a lot of functionality
9. errors are missing context
	* the logging middlewares are quite limited in what information they can share. This requires

</details>



<details><summary>References</summary>

16. [Avoid Package Names like base/util/common - Dave Cheney](https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common)

</details>

### References [[top]](#how-to-inherit-a-mess)

1. [How I write HTTP Services After 8 Years - Matt Ryer](https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html) [[section](#2-add-implementation-of-allsrv-server-top)]
2. [Concurrency in Go - Katherine Cox-Buday](https://katherine.cox-buday.com/concurrency-in-go/) [[section](#19-serialize-access-for-in-mem-value-access-to-rm-race-condition-top)]
3. [Intro to Versioning a Rest API](https://www.freecodecamp.org/news/how-to-version-a-rest-api/) [[section](#22-add-v2-server-create-api-top)]
4. [Versioning Rest Web API Best Practices - MSFT](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design#versioning-a-restful-web-api) [[section](#22-add-v2-server-create-api-top)]
5. [API Design Cheat Sheet](https://github.com/RestCheatSheet/api-cheat-sheet#api-design-cheat-sheet) [[section](#22-add-v2-server-create-api-top)]
6. [InfluxDB Testing](https://github.com/influxdata/influxdb/tree/v2.0.0/testing) [[section](#27-add-db-test-suite-top)]
7. [Vice Test Suite](https://github.com/matryer/vice/blob/master/vicetest/test.go) [[section](#27-add-db-test-suite-top)]
8. [pprof tool package](https://pkg.go.dev/runtime/pprof) [[section](#33-add-pprof-routes-top)]
9. [pprof HTTP integration](https://pkg.go.dev/net/http/pprof) [[section](#33-add-pprof-routes-top)]
10. [Profiling Go Programs - Russ Cox](https://go.dev/blog/pprof) [[section](#33-add-pprof-routes-top)]
11. [errors module](https://github.com/jsteenb2/errors) [[section](#34-add-githubcomjsteenb2errors-module-to-improve-error-handling-top)]
12. [SDK module - github.com/jsteenb2/allsrvc](https://github.com/jsteenb2/allsrvc) [[section](#35-add-githubcomjsteenb2allsrvc-sdk-module-top)]
13. [Setting version in SDK via debug.BuildInfo](https://github.com/jsteenb2/allsrvc/blob/main/client.go#L21-L30) [[section](#35-add-githubcomjsteenb2allsrvc-sdk-module-top)]
14. [Perseus module tracker](https://github.com/CrowdStrike/perseus) [[section](#35-add-githubcomjsteenb2allsrvc-sdk-module-top)]
15. [Threedots Tech Introduction to Clean Architecture](https://threedots.tech/post/introducing-clean-architecture/) [[section](#36-add-slimmed-down-wild-workouts-go-ddd-example-repo-top)]
16. [Avoid Package Names like base/util/common - Dave Cheney](https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common) [[section](#37-add-thoughts-on-wild-workouts-clean-architecture-top)]


### Suggested Resources [[top]](#how-to-inherit-a-mess)

1. [TDD, Where Did It All Go Wrong - Ian Cooper](https://www.youtube.com/watch?v=EZ05e7EMOLM)
    * Excellent talk that touches on some busted testing paradigms and how to remedy them
2. [Absolute Unit (Test) - Dave Cheney](https://www.youtube.com/watch?v=UKe5sX1dZ0k)
    * Builds on Ian's talk, but provides it in the context of `go` with useful examples
3. [Intentional Code - Minimalism in a World of Dogmatic Design - David Whitney](https://www.youtube.com/watch?v=8j4fhsLcT4k)
4. [Error Handling in Upspin - Rob Pike](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html)
    * Incredibly insightful look at what you can do for errors. This was largest inspiration for the [errors](https://github.com/jsteenb2/errors) module
5. [Generics Unconstrained! - Roger Peppe](https://www.youtube.com/watch?v=eU-w2psAvdA)
    * The best resource for `go` generics
6. [Systems Thinking Demystified - Dr. Russ Ackoff](https://www.youtube.com/watch?v=OqEeIG8aPPk)
    * One of the most influential disciplines I've ever picked up, systems thinking will improve your engineering
7. [0-60 Systems Thinking Resources](https://gist.github.com/jsteenb2#systems-thinking---go-beyond-the-jira-ticket)
    * Ton of useful resources for picking up systems thinking `:YAAAAAS:`
8. [Clean Up Your GOOOP: How to Break OOP Muscle Memory - Dylan Bourque](https://www.youtube.com/watch?v=qeTzjeuq3cw)
    * Deep dive into OOP patterns that show up in `go` and how to make sense of it all
9. [High-Assurance Go Cryptography - Filippo Valsorda](https://www.youtube.com/watch?v=lkEH3V3PkS0)
    * Intriguing look at how the `go` crypto contributors use testing to maintain velocity, safety and endurance

---

generated by [github.com/jsteenb2/gitempl](https://github.com/jsteenb2/gitempl)