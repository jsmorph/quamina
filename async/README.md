# Experiment: background rebuilding

## Motivation

1.  In the core Matcher implementation, locking primitives to support
    concurrent `AddPattern` and `Matches` operations result in
    noticeable overhead.

1.  The Pruner exists because the core Matcher doesn't support
    `DeletePattern`.

Tim suggested the following approach (maybe!) that trades those
complexities for some lag in when `AddPattern` and `DeletePattern`
operations take effect.

## Design

1.  Accumulate `AddPattern` and `DeletePattern` mutations but do not
    apply them immediately.

1.  Periodically rebuild the enter core Matcher from scratch in the
    background.  Atomically update the core Matcher in use after each
    rebuild.

## Discussion

This design eliminates the need for primitive locking in the core
Matcher.  This design also eliminates the need for the Pruner.

The costs are:

1.  Mutation lag: (`AddPattern` and `DeletePattern`) do not take
    effect immediately.  There's a lag (which can have configurable
    *minimum*).  The maximum lag is determined by the number of live
    patterns and their complexity (i.e., rebuild time).

1.  RAM: The implementation maintains two complete core Matchers and a
    set of live patterns (and some other stuff).

1.  CPU: The core Matcher is continuously rebuilt from scratch rather
    than being mutated in place.  (The Pruner does that, so this
    approach is perhaps not that much worse in that respect.)
