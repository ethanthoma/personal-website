# Odin Errdefer

I love [Zig](https://ziglang.org/). It's simple, it's nice, and I am still 
salty about [no comptime pointers](https://github.com/ziglang/zig/issues/7396). 

But still, it's my go-to language for unmanaged memory. Why? Because other 
languages do too much (cpp, rust, whateva else) and the larger the surface area,
the longer to learn the gotchas; simply a bigger surface area for failure.

I'm lazy. I want a good STD, clear/simple syntax, and errors-as-values. So 
basically, I want a better C (on the managed-memory side of the world).

But Zig isn't the only racehorse in replacing C, there's [Odin](https://odin-lang.org/).

## Magic of `errdefer`

One of Zig's most banger features is its `defer` and `errdefer` statements. 
This lets you run *things* at the end of the current scope. Take this Zig 
function, for instance:

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

### But What About Odin

Odin doesn't have `errdefer`, only `defer`. In fact, it's not really possible 
for Odin to have `errdefer` because Odin does even have an error type.

In Zig, errors are a special kind of type: a global, modifiable error enum. This 
lets the Ziglang devs make specialized syntax just for the error type. 

Odin, instead has two idiomatic ways to manage errors: enums or bools. 

Here's how we can emulate `errdefer` in Odin using `defer` and an `enum`:

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

The trick of this trick is:
- By including `err` in the function signature, we have a reference to the error 
    state within our `defer` statement.
- The `defer` is conditional, checking the value of `err` before deciding 
    whether to clean up `copied_slice`.
- By default, `err` is `.None`, signaling no error.

### It was Trap: Default Values

Odin has the concept of default values. When you define a variable without a 
value, it gets set to a default value:

```odin
some_number: int // this is equal to 0
```

For booleans, the default is `false` which makes sense (from the C world) but 
leads to a sneaky bug if you're not careful:

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

We tried to be smart here by skipping the loop if `find` and `replace` are the 
same. But if they are the same, `ok` remains `false` (the default), and our 
`defer` goes ahead and deletes the `copied_slice` memory and then returns it, 
making a dangling pointer. Oops.

### The Fix: Assume Success

To dodge this pitfall, we have to explicitly set `ok` to `true`. In other words,
like every other function, assume success unless otherwise:

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

It is a little bit annoying as (1) the STD doesn't do this so it's hard to be 
aware of this pitfall till you fall and (2) default success is true for the enum
error and no error, only defaults to failure for bools which is bad design imo.

But I think this pitfalls shows the real difference between Odin and Zig. In Zig,
everything is explicit. There is no [default allocator](https://odin-lang.org/docs/overview/#allocators) 
and there are no default values. You will always know, explicitly, what things 
are. 

This is why I love Zig. And also why Zig removed my beloved comptime vars, as 
the execution order was hard to parse. Zig's manifesto of making the language 
readable and simple in syntactic sugar aligns with what I believe about coding. 
Programming languages should be simple. 

And I love Odin for that. Sure, it's not as explicit as Zig but it is simple, 
it is readable, and there's even less special syntax and sugar and jazz than Zig.
Like errors for example. In Odin, it is just a normal type like anyother.

So try Odin. It is probably the best C-style language for graphics or for small 
systems projects. I'd still use Zig (or maybe Rust one day) for big projects, 
like my [tensor library](https://github.com/ethanthoma/zensor), but I think 
Odin is perfect for anything else. *Lucky coding*.
