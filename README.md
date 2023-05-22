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

BuildX
------
```
go run bin/make.go setup
go run bin/make.go buildx
```

Style
-------------
Deliberately, the API may seem terse, but it should afford smooth reading animation sequences. Normally, I prefer fully spelled out variables that emphasize data hierarchies and fn names in object-verb order. eg. orderItemPut(item: Item). Think "dot-notation" without the dots.

I feel this strategy helps...
  * control entropy since I only need to "think" in one order (dot-order).
  * improve maintainability... because remembering the abbreviation over time becomes difficult.

I like to avoid use of the shift key, so snake_case and PascalCase are right out for me. Kebab-case... not bad, but I prefer camelCase to keep things tight, but prefer opinionated languages like GoLang that eliminate the conversation with strictly enforced style.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.1.png)

Ports
-----------------------
### GoLang
#### Re: Building for Web and WASM
Go compiles to multiple platforms and architectures and allows binding to platform C libs with the `CGO_ENABLED=true` build flag. GAS for GoLang leverages this [staticly included SDL Libs for GoLang](https://github.com/veandco/go-sdl2) to build custom executables for MacOS, Linux/POSIX, and windows containers of various architectures (arm, i386, etc.).

But getting GAS in GoLang to run in a javascript or web container getsc complicated.

GoLang can compile *directly to* javascript (no intermediate binary) with a completely separate toolchain (see  GopherJS below). But to run anything *in* javascript or a browser (i.e. embedded), we need to build out to WebAssembly (WASM). While WASM clearly outperforms native javascript in Apples to Apples comparison, I wonder how code that relies on many calls to native browser APIs compares. As I understand it, WASM calls through to .js first which ultimately calls to native browser APIs. How this call chain works under the hood could have a big impact on WASM performance, though I'm sure WASM will get better access to these APIs over time.

##### Native Toolchain
The native GoLang toolchain can build a WASM target using something like `GOOS=js GOARCH=wasm go build -o main.wasm`. But it does not support binding to C libs for the wasm target, since a WASM app runs in a sandbox that does not and cannot provide those C dependencies.

Even if the native toolchain could generate a build with staticly included C libs, only *portable* c libs without additional dependencies would run. Since the Go maintainers have not promised this support, you will need to port your portable C libs to GoLang.

##### Emscripten
Googling for "GoLang SDL WASM" surfaces [EMScripten](https://emscripten.org/docs/introducing_emscripten/about_emscripten.html), a toolchain for compiling various languages to WASM and asm.js (a tiny, high-performance subset of .js).

It has similar limitations. It can only build *portable* code and libraries to WASM targets as described in [Emscripten Porting](https://emscripten.org/docs/porting/guidelines/index.html).  And yet... Emscripten *does* provide support for building applications that use the SDL2 CLibs. How?

The Emscripten toolchain replaces calls to C SDL libs with WASM calls to a .js port of those SDL libs that call browser-native WebGL. Hacky, but it’s "Good Enough" such that older games just might work. See [*Javascript* port of SDL1.2](https://github.com/emscripten-core/emscripten/blob/main/src/library_sdl.js).

For SDL2, the Emscripten toolchain ports the SDL2 codebase to the target, so you can literally just compile it and link it to your project. See [SDL and WebAssembly](https://discourse.libsdl.org/t/sdl-and-webassembly/24611/5). Emscripten already has the smarts to convert OpenGL calls to WebGL calls, so without looking deeper, I'll surmise that's what happens here.

See this [Comparison of SDL1.2 and SDL2 with Emscripten](https://www.jamesfmackenzie.com/2019/12/01/webassembly-graphics-with-sdl/)


##### Emscripten and GoLang
Emscripten does not support GoLang. It only compiles languages that use LLVM compiler/linker toolchains, but GoLang has a custom compiler/linker toolchain. Though [GoLLVM](https://go.googlesource.com/gollvm/) could provide that compatibility some day, it is not production-ready.

See also...
* [WASM Spec Summary](https://webassembly.github.io/spec/)
* [Mozilla Intro to WASM Concepts](https://developer.mozilla.org/en-US/docs/WebAssembly/Concepts)
* [The WASM spec](https://github.com/WebAssembly/spec/)
* [Using WebAssembly to call Web API methods](https://stackoverflow.com/questions/40904053/using-webassembly-to-call-web-api-methods).

##### GopherJS
GopherJS compiles GoLang code to .js with a completely separate toolchain. It does not create a binary executable, not even as an intermediate. See [Creating WebGL apps with Go](https://blog.gopheracademy.com/advent-2018/go-webgl/) for a story about using GopherJS with three.js to do 3D rendering.

While people rave about GopherJS, your Go code will never look the same. You will essentially write Javascript in Go and the Go to .js binding code won't compile to a GoLang binary. So your GoLang code will *only* compile to .js. If you have the option, you're probably better off writing Typescript IMO. WASM intends to one day replace .js and improve its performance, so why target .js at all?

By itself, WebAssembly cannot directly call Browser Native APIs, like DOM or WebGL; it can only call JavaScript, passing in integer and floating point primitive data types. So, if a WASM app needs to call Web APIs frequently, it goes through .js, which can impact performance.

Putting a finer point on it... GopherJS provides a binding interface and [bindings to popular .js packages](https://github.com/gopherjs/gopherjs/wiki/bindings). When GopherJS compiles the GoLang code it produces .js that calls .js through the generic binding interface.

##### Back to the native WASM target
If you build to WASM with the native GoLang toolchain, your GoLang to .js bindings flow through the generic GoLang to .js binding interface... [syscall/js](https://cs.opensource.google/go/go/+/master:src/syscall/fs_js.go) that essentially implements communication according to the WASM spec.

The picture shaping up here... WASM can optimize compute intensive operations, but might not provide performance advantage when used to drive the display layer, which requires high frequency WebGL calls that must flow through .js as a broker.

##### The Future
If/when WASM apps get direct access to WebAPIs that ultimately bypass .js... the table flips. [WASI](https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-intro.md), the Web Assembly System Interface, represents some progress in this direction. TinyGo leads the way in WASI support for GoLang. See [WASI Hello World](https://wasmbyexample.dev/examples/wasi-hello-world/wasi-hello-world.go.en-us.html).

Whether you compile your GoLang code to .js with GopherJS or WASM with the native toolchain, your GoLang code will call out of the container through some sort of bindings layer.  Those bindings need to be written by a person or generated by a compiler.

So far, it seems that most of the energy has gone into the development of Gopher.js bindings (https://github.com/gopherjs/gopherjs/wiki/bindings), but one notable developer ported his [GopherJS Bindings for Three.js](https://github.com/soypat/gthree) to [syscall/js Bindings for Three.js](https://github.com/soypat/three), making it possible to use Three.js in GoLang to build WASM.

Obviously three.js doesn't provide everything that SDL provides... no abstrations for windows and UI, etc. But the Web API provides all of this and occupies roughly the same tier in the front-end stack. Why not use the Web API?

Ultimately your Game needs to drive a Game Engine for rendering, sound, and input processing that works in web, desktop, and mobile containers. And if you plan to build a stack, you might as well use something that supports 3D. And, ideally, it is positioned to take advantage of the future of rendering, [WebGPU](https://en.wikipedia.org/wiki/WebGPU). See also [go-webgpu](https://github.com/rajveermalviya/go-webgpu).

Unless an existing GoLang project has solved this problem, the fast path might involve porting three.js to GoLang with calls to OpenGL which easily bind to syscall/js and WASM.

Fortunately, [g3n](https://github.com/g3n/engine) seems to have solved these problems and has some momentum. More to come...



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
* Crashing
  The binary crashes periodically and I've seen mention that CPU / Thread pinning might solve the problem. Overall, I'm leaning away from SDL as a solution for KVM access and toward something like (Go-WebGPU)[https://github.com/rajveermalviya/go-webgpu]


License
-------
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.




See Also
--------
* [g3n](https://github.com/g3n/engine), a GoLang 3D game engine that may actually generate WASM targets.
* [Mach Engine](https://machengine.org/), a progressive Game Engine in Zig written by an obsessed maniac.
* [Zig](https://ziglang.org/), a general-purpose langauge in the family of Rust and Carbon intended to succeed C.
* [The Future of Graphics with Zig](https://devlog.hexops.com/2021/mach-engine-the-future-of-graphics-with-zig/),
* [Three.js GoLang Example] https://github.com/soypat/threejs-golang-example
