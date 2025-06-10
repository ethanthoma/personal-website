# Odin Errdefer

I love [Zig](https://ziglang.org/). The language is simple and clear. Sure it
isn't the easiest for noobies to read but its explicitness is its appeal.

While I'm still salty about the
[removal of comptime pointers](https://github.com/ziglang/zig/issues/7396), Zig
is still my go-to language for manual memory management.

Why?

Because many other languages, C++ and Rust, do too much and have a lot of
"gotchas" you just have to know. I like smaller surface areas, smaller chances
for failure.

My needs are pretty straightforward. I want a good STD, clear/simple syntax, and
errors-as-values (theses have changed a bit but shhh). basically, I'm looking
for a modern C. And to be honest, I think of Zig as a more C++ alternative. For
the C world, there's [Odin](https://odin-lang.org/).

## Magic of `errdefer`

One of Zig's most banger features is its `defer` and `errdefer` statements. This
lets you run *things* at the end of the current scope. Take this Zig function,
for instance:

```zig
fn find_and_replace(
    allocator: std.mem.Allocator,
    slice: []const u32,
    find: u32,
    replace: u32,
) ![]u32 {
    var copied_slice = try allocator.dupe(u32, slice)
    errdefer allocator.free(copied_slice)

    var find_is_in_slice = false
    for (slice, 0..) |elem, index| {
        if (elem == find) {
            find_is_in_slice = true
            copied_slice[index] = replace
        }
    }

    if (!find_is_in_slice) {
        return error.FindIsNotInSlice
    }

    return copied_slice
}
```

The gist is

- We want to find all ints with value `find` in our `slice` and replace them
  with `replace`, if there are no replacements, return an error.
- We dup(licat)e the input `slice`, using the passed in `allocater`.
- We return either the `copied_slice` or an error.
- The `errdefer` ensures that if an error pops up, `copied_slice` is freed
  automatically, preventing memory leaks.

### What About Odin

Odin doesn't have `errdefer`, only `defer`. In fact, it's not really possible
for Odin to have `errdefer` because Odin doesn't even have a special error type.

In Zig, errors are a dedicated type: a global, modifiable error enum. Instead,
Odin, has two idiomatic ways to manage errors: enums or bools.

Here's how we can emulate `errdefer` in Odin using `defer` with an `enum`:

```odin
Error :: enum {
    None,
    FindIsNotInSlice,
}

find_and_replace :: proc(slice: []u32, find, replace: u32) -> (err: Error) {
    copied_slice := make([]u32, len(slice))
    defer if err != .None do delete(copied_slice)
    copy(copied_slice, slice)

    find_is_in_slice: bool = false
    for &elem, index in slice {
        if elem == find {
            find_is_in_slice = true
            copied_slice[index] = replace
        }
    }

    if !find_is_in_slice do return .FindIsNotInSlice

    return err
}
```

The trick here is:

- By including `err` in the function signature, we have a reference to the error
  state within our `defer` statement.
- The `defer` is conditional, checking the value of `err` before deciding
  whether to clean up `copied_slice`.
- By default, `err` is `.None`, signaling no error.

### It was Trap: Default Values

Odin has a feature of setting default values for uninitialized variables. For
instance, an int defaults to 0:

```odin
some_number: int // this is equal to 0
```

For booleans, the default is `false` which makes sense (from the C world) but
leads to a sneaky bug if you're not careful.

Consider this alternative implementation using a boolean to signal success:

```odin
find_and_replace :: proc(slice: []u32, find, replace: u32) -> (ok: bool) {
    copied_slice := make([]u32, len(slice))
    defer if !ok do delete(copied_slice)
    copy(copied_slice, slice)

    if find != replace {
        for &elem, index in slice {
            if elem == find {
                ok = true
                copied_slice[index] = replace
            }
        }
    }

    return ok
}
```

We tried to be smart here and skip the loop if `find` and `replace` are the same
value. But, if they are the same, `ok` never gets set to `true` and stays its
default `false` value. Our `defer` statement then goes ahead and deletes the
`copied_slice` memory before it's returned, making it into a dangling pointer.
Oops.

### The Fix: Assume Success

To dodge this pitfall, we must be explicit and assume the function will succeed:

```odin
find_and_replace :: proc(slice: []u32, find, replace: u32) -> (ok: bool = true) {
    copied_slice := make([]u32, len(slice))
    defer if !ok do delete(copied_slice)
    copy(copied_slice, slice)

    if find != replace {
        for &elem, index in slice {
            if elem == find {
                ok = true
                copied_slice[index] = replace
            }
        }
    }

    return ok
}
```

By setting the default value of `ok` to `true` in the function signature, we
_assume success_.

This is a little bit annoying. First, the STD doesn't set boolean errors as
`true` by default so it's hard to be aware of this pattern till it bites you in
the ass. Secondly, functions with no errors or use enums for error assume the
function was successful. It is only for booleans where the default assumes
failure, it's more the inconsistency that makes this an issue.

## A More Philosophical Difference

Odin and Zig both heavily emphasize explicitness, just differently.

In Zig, everything is explicit. You always know what things are. It doesn't have
a [default allocator](https://odin-lang.org/docs/overview/#allocators) and there
are no default values. It's why Zig is so readable and why why features that
obscured execution flow, like my beloved comptime vars, were removed. And this
aligns with what I believey about coding:

**Programming languages should be simple.**

And I don't mean simple like golang (ew). I mean the state-machine a coder keeps
in their head should be as simple as possible. It should be the compilers job,
not mine.

Odin is a bit different from Zig in this sense. While it is also simple and
readable, Odin makes a few more concessions for convenience. There's less
special syntax, sugar, and jazz than Zig.

Yet, we can still make `errdefer`.

I love languages that can get away with using one feature for many. Like gleam's
`use` for early returns. It's that smaller surface error that makes **me less
likely to make mistakes**.

So, give Odin a try. It is probably the best C-style language that is simple and
a pleasure to use. *Lucky coding!*
