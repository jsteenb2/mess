// package wild_workouts is an introductory repository that provides a working example
// of working with clean architecture in go.
//
//Praises:
// * "Clean" code, good separation of concerns, i.e. no http code is driving the app or repo
//   implementation.
//		* I like "Clean" code. But I like to fit the design patterns to the language at hand as well. A
//	 	  good chunk of "Clean" is necessary b/c of how overly complex class oriented languages make
//	 	  modeling your problem space. "Clean" is an attempt at wrangling this kraken so that a team can
//	 	  maintain and extend it further. If you can take one thing away from "Clean"/"Hexagonal"/"Onion"
//		  it's this, empower your design with encapsulation/separation of concerns.
// * gives the system/codebase structure
//		* Without an intended design for your system/codebase's structure, you'll end up with a giant
//		  mess on your hands. We'll explore this in our next example.
//	* familiar to folks coming from another language to go
//		* b/c it isn't idiomatic go, and looks and feels like Java, it can be more comfortable
//		  for folks coming to go from various OO languages that stress "Clean" or similar.
//	* provides orthogonal concerns via middleware
//		* though, one word of caution on logging middlewares. If _everything_ gets a logging middleware
//		  it'll make your logging infrastructure VERY noisy. I don't find a ton of value in logging at
//		  all layers of a service. I find it more burdensome than helpful most the time. Having errors
//		  that provide context, can improve your logging experience DRAMATICALLY.
// * tests at the feature level, this is fantastic.
// * errors are structured
//		* can't say enough good things about structured error handling. Adds a ton of value to
//		  any and all service codebases.
//
//
//Concerns:
// * whiplash, do we need a ton of pkgs to make code "Clean"?
//		* spoiler alert, you do not :-)
//	* TOO much "structure"
//		* hot take: this is the reason many go devs snear at "clean" go code, b/c its actually
//		  			messy in practice. With Go's pkg oriented design we don't need to separate by "what is"
//					but rather... what it's trying to accomplish/do. This "Clean" style does wonders for
//					getting	consultancies paid... but leads to the inevitable mess with any mixed seniority
//					group.
//					DC did a good writeup on pkg naming here:
//						https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common
//		* take a count, how many layers of pkgs do you have to jump through to understand how
//		  something works?
//		* how will a younger more green engineer fair?
//	* codebase requires an uphill climb to understand how to work with the "Clean" principles
//		* Newer engineers will have an uphill climb to learn the ways of your companies "Clean" implementation.
//		  For senior engineers... they'll have worked with "Clean" in one way or another. The eventual
//		  bikeshed moments will transpire allowing seniors to debate at ends about meaningless things.
//		  These are the two extremes I've run into with "Clean", and I am willing to wager I'm not the
//	 	  only one who has experienced this.
//	* a great example of writing Java code with Go
//		* for experienced go devs, this can be extremely draining. The other big downside is...
//		  these patterns will "partially" reproduce as the company grows. I say partially,
//		  b/c the patterns will become copy pastaed all over with each team adding their own unique
//		  idiosyncratic takes.
//	* modules for :allthethings:, when you have the following line in your module, you are
//	  likely "holding it wrong":
//		replace github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common => ../common/
//		* The boundaries of modules in this repo don't make sense to me (if you understand it,
//		  please explain it to me). This one really gives me chills as the entire module within
//	 	  each application (trainer/training/common) are exported.... meaning hyrum's law is
//	 	  just around the corner... :cringe:
//	* common module
//		* this harkens back to the issue before where we're using principles that work in other
//		  languages, but languish in go. For example, take the decorators pkg... why do we need
//		  `decorator.ApplyCommandDecorators` in its own pkg???
//		  Contrast the indirection with our first example, "allsrv", where everything is fairly close
//		  in proximity and implementation. Its radically simpler to juggle in an average developer
//		  like me's head. Don't require a 200 IQ to contribute.
//		* Why do we need a ApplyCommandDecorators in the common/metrics pkg instead of keeping
//		  this encapsulated inside the app/command pkg?
// * tests are missing for a lot of functionality
//	* errors are missing context
//		* the logging middlewares are quite limited in what information they can share. This requires
//		  a bit more digging to rebuild the failure.
//
//STORY TIME: The tale of Abu Fawwaz

package wild_workouts
