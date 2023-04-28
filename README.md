GAS: Game Animation System
==========================
The game animation system (GAS) animates objects for interactive media of all sorts including video games and other motion graphics.

I intend to translate GAS into multiple languages as a learning exercise.  Right now, I only provide GoLang.

I would not recommend GAS herein for production. Nevertheless, this repo can serve as a good place to start and extend. Custom animation systems can often do more with less, so you might want to experiment.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.2.png)

Documentation
-------------
No formal documentation other than comments, source code, and the example Frogger Welcome Screen.

The API mostly explains itself, though inspection of the example code will yield the most rapid understanding.


Installation
------------
Pre-Requisites:
See https://github.com/veandco/go-sdl2

Code:
```
git clone https://github.com/jkassis/gas
cd gas/go
go run main.go
```

Style
-------------
Deliberately, the API may seem terse, but it should afford smooth reading animation sequences. Normally, I prefer fully spelled out variables that emphasize data hierarchies and fn names in object-verb order. eg. orderItemPut(item: Item). Think "dot-notation" without the dots.

I feel this strategy helps...
  * control entropy since I only need to "think" in one order (dot-order).
  * improve maintainability... because remembering the abbreviation over time becomes difficult.

I like to avoid use of the shift key, so snake_case and PascalCase are right out for me. Kebab-case... not bad, but I prefer camelCase to keep things tight.

That said, I'm flexible to whatever the team chooses and prefer opinionated languages like GoLang that strictly enforce style.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.1.png)

Ports
-----------------------
### GoLang
#### Re: Building for Web and WASM
Go compiles to multiple platforms and architectures and allows binding to platform C libs with the `CGO_ENABLED=true` build flag. GAS for GoLang leverages this [staticly included SDL Libs for GoLang](https://github.com/veandco/go-sdl2) to build custom executables for MacOS, Linux/POSIX, and windows containers of various architectures (arm, i386, etc.).

But getting GAS in GoLang to run in a javascript or web container getsc complicated.

GoLang can compile *to* javascript with a completely separate toolchain (see  GopherJS below). But to run *in* javascript or a browser, we need to build out to WebAssembly (WASM).

##### Native Toolchain  
The native GoLang toolchain can build a WASM target using something like `GOOS=js GOARCH=wasm go build -o main.wasm`. But it does not support binding to C libs for the wasm target, since a WASM app runs in a sandbox that does not and cannot provide them.

Even if the native toolchain could generate a build with staticly included C libs, only *portable* c libs without additional dependencies would run. Since the Go maintainers have not promised this support, you will need to port your portable C libs to GoLang.

##### Emscripten  
Googling for "GoLang SDL WASM" surfaces [EMScripten](https://emscripten.org/docs/introducing_emscripten/about_emscripten.html), a toolchain for compiling various languages to WASM and asm.js (a tiny, high-performance subset of .js).

It has similar limitations. It can only build *portable* code and libraries to WASM targets as described in [Emscripten Porting](https://emscripten.org/docs/porting/guidelines/index.html).  And yet... Emscripten *does* provide support for building applications that use the SDL2 CLibs. How?

The Emscripten toolchain replaces calls to C SDL libs with WASM calls to a .js port of those SDL libs that renders with WebGL. Hacky, but it’s "Good Enough" such that older games just might work. See [*Javascript* port of SDL1.2](https://github.com/emscripten-core/emscripten/blob/main/src/library_sdl.js).

For SDL2, the Emscripten toolchain ports the SDL2 codebase to the target, so you can literally just compile it and link it to your project. See [SDL and WebAssembly](https://discourse.libsdl.org/t/sdl-and-webassembly/24611/5). Emscripten already has the smarts to convert OpenGL calls to WebGL calls, so without looking deeper, I'll surmise that's what happens here.

See this [Comparison of SDL1.2 and SDL2 with Emscripten](https://www.jamesfmackenzie.com/2019/12/01/webassembly-graphics-with-sdl/)


##### Emscripten and GoLang  
Emscripten only compiles LLVM-based languages. GoLang has a custom compiler toolchain. And [GoLLVM](https://go.googlesource.com/gollvm/) is not production-ready. See also...
* [WASM Spec Summary](https://webassembly.github.io/spec/)
* [Mozilla Intro to WASM Concepts](https://developer.mozilla.org/en-US/docs/WebAssembly/Concepts)
* [The WASM spec](https://github.com/WebAssembly/spec/)
* [Using WebAssembly to call Web API methods](https://stackoverflow.com/questions/40904053/using-webassembly-to-call-web-api-methods).

##### GopherJS  
GopherJS compiles GoLang code to .js with a completely separate toolchain. It does not create a binary executable, not even as an intermediate. See [Creating WebGL apps with Go](https://blog.gopheracademy.com/advent-2018/go-webgl/) for a story about using GopherJS with three.js to do 3D rendering.

While people rave about GopherJS, your Go code will never look the same. You will write Javascript in Go and the Go to .js binding code won't compile to a GoLang binary. If you have the option, you're probably better off writing Typescript IMO. WASM intends to improve the performance .js anyway, so why target .js?

By itself, WebAssembly cannot directly access the DOM or WebGL; it can only call JavaScript, passing in integer and floating point primitive data types. So, if WASM app needs to call Web APIs frequently, it goes through .js, which can impact performance.

Putting a finer point on it... GopherJS provides a binding interface and [bindings to popular .js packages](https://github.com/gopherjs/gopherjs/wiki/bindings). When GopherJS compiles the GoLang code it produces .js that calls .js through the generic binding interface.

##### Back to the native WASM target  
If you build to WASM with the native GoLang toolchain, your GoLang to .js bindings flow through the generic GoLang to .js binding interface... [syscall/js](https://cs.opensource.google/go/go/+/master:src/syscall/fs_js.go) that essentially implements communication according to the WASM spec.

The picture shaping up here... WASM can optimize compute intensive operations, but might not provide performance advantage when used to drive the display layer, which requires high frequency WebGL calls that must flow through / originate from .js.

##### The Future  
But if/when WASM apps get direct access to WebAPIs that ultimately bypass .js... the table flips. [WASI](https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-intro.md), the Web Assembly System Interface, represents some progress in this direction. TinyGo leads the way in WASI support for GoLang. See [WASI Hello World](https://wasmbyexample.dev/examples/wasi-hello-world/wasi-hello-world.go.en-us.html).

Whether you compile your GoLang code to .js with GopherJS or WASM with the native toolchain, your GoLang code will call out of the container through some sort of bindings layer.  Those bindings need to be written by a person or generated by a compiler.

So far, it seems that most of the energy has gone into the development of Gopher.js bindings (https://github.com/gopherjs/gopherjs/wiki/bindings), but one notable developer ported his [GopherJS Bindings for Three.js](https://github.com/soypat/gthree) to [syscall/js Bindings for Three.js](https://github.com/soypat/three), making it possible to use Three.js in a WASM build.

It isn't SDL, but perhaps we don't need SDL for the Web? It essentially defines abstrations for rendering, sounds, windows, etc. The Web API provides all of this. So the Web API and the SDL API occupy roughly the same tier in the front-end stack. Why not use the Web API?

At this point, it comes down to convenience, time, and abstractions. Your Game needs to operate a Game Engine that works in web, desktop, and mobile containers. It will take some glue to make it work. Even if you choose three.js as the interface to the view, some port of three.js needs to exist for GoLang. To continue... roll up sleeves.


#### Known Issues
* Datatypes
  excessive casting and use of int64 creating a large memory footprint for display objects.
* Error Handling
  The SDL layer surfaces many errors which GAS mostly discards. Once again... not production ready.
* Public API
  Little thought into what fields should allow public access
* Recycling
  This systems should recycle dobs to avoid garbage collection. Currently no pooling of textures, dobs, etc.
* z-layering
  Controlled by OrderedMaps, which I have not benchmarked, tested thoroughly.
* fn calls
  Ans implement an abstract interface so that Dobs can run generic Ans. This means we have to access properties through accessors. Should be possible to fix this with generics and inverting the embed. Instead of embedding BaseAnim in each An, we would embed the specific An inside base.
  It should also be possible with this fix to init startTick for each An before the first invocation of the start tick to remove some of the boilerplate code from the An.Tick fns.
* memory leak
  looks like there is a huge memory leak that i need to shake down.


License
-------
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.




See Also
--------

### GoLang for Cross-Platform Game Engines
https://blog.gopheracademy.com/advent-2018/go-webgl/
https://github.com/soypat/threejs-golang-example
