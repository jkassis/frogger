GAS: Game Animation System
==========================
The game animation system (GAS) animates objects for interactive media of all sorts including video games and other motion graphics.

I intend to translate GAS into multiple languages as a learning exercise.  Right now, only GoLang is provided.

I would not recommend GAS herein for production. Nevertheless, this repo can serve as a good place to start and extend. Custom animation systems can often do more with less, so you might want to experiment.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.2.png)

Documentation
-------------
No formal documentation other than comments, source code, and the example Frogger Game Welcome Screen.

The API mostly explains itself, though inspection of the example Frogger anim code will yield the most rapid understanding.


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

Known Issues
-----------------------

GoLang
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


## License
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.



