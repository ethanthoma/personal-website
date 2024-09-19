# Odin Errdefer

If you know me (IRL or xitter), I love [Zig](https://ziglang.org/). It's simple, 
it's nice, and [no comptime pointers](https://github.com/ziglang/zig/issues/7396) 
make me depressed. But still, it's my go-to language for unmanaged memory. Why? 
Because other languages do too much (cpp, rust, whateva else). I'm lazy. I want 
a good STD, clear/simple syntax, and errors-as-values.

So basically, a better C (cpp is clearly the opposite).

But Zig isn't the only racehorse in the C replacement race: there's [Odin](https://odin-lang.org/).

## Magic of `errdefer`

One of Zig's most enticing features is its `defer` and `errdefer` statements. 
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

For the none zigmas, the gist is
- **Purpose**: we want to find all ints with value `find` in our `slice` and 
    replace them with `replace`, if there are no replacements, return an error
- **Memory Allocation**: we dup(licat)e the input `slice`, using the passed in 
    `allocater`
- **Error Handling**: we return either the `copied_slice` or an error
- **Automatic Cleanup**: the `errdefer` ensures that if an error pops up, 
    `copied_slice` is freed automatically, preventing any memory leaks

### But What About Odin

Odin doesn't have `errdefer`, only `defer`. In fact, it's not really possible 
for Odin to have `errdefer` because Odin does even have an error type.

In Zig, errors are a special kind of type: a global, modifiable error enum. This 
lets the Ziglang devs make specialized syntax. Instead, in Odin, there are two 
idiomatic ways to manage errors: enums or bools. 

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

- **Explicit Output Parameter**: by including `err` in the function signature, 
    we have a reference to the error state within our `defer` statement
- **Conditional Defer**: the `defer` checks the value of `err` before deciding 
    whether to clean up `copied_slice`
- **Default Value**: by default, `err` is `.None`, signaling no error

### It was Trap: default values

Odin has the concept of default values. When you define a variable without a 
value, it gets set to a default value:

```odin
some_number: int // this is equal to 0
```

For booleans, the default is `false` which makes sense but can lead to a sneaky 
bug if you're not careful. Consider this function:

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
making a dangling pointer.

Why? Because the default value of a bool is `false` and since we didn't 
explicitly set `ok` to true then we get...well...a false.

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

And so everything works now. But I think that the problem I showed here is the 
differential between Odin and Zig. In Zig, everything is explicit. There is no 
[default allocator](https://odin-lang.org/docs/overview/#allocators) and there 
are no default values, so you can know, explicitly, what things are and what 
things will do. This is the same (well, one of) reason Zig removed my beloved 
comptime vars, of which I'm still salty about, as the execution order was hard 
to parse. 

I love Zig's manifesto of making the language readable and simple in syntactic 
sugar. Programming languages should be simple. And I love Odin for that. Sure, 
it's not as explicit as Zig but it is simple, it is readable, there's even less 
special syntax and sugar and jazz than Zig.

So try Odin. I'd still probably use Zig in a big project, like my [tensor library](https://github.com/ethanthoma/zensor),
but I think Odin is perfect for anything else. Happy coding.
