GAS: Game Animation System
==========================
The game animation system (GAS) animates objects for interactive media of all sorts including video games and other motion graphics.


The author's intent is to translate GAS into multiple languages as a learning exercise.  Right now, only GoLang is provided.

GAS herein is not production ready. Nevertheless, it is a place to start and extend and custom animations systems can often do more with less, so you might want to experiment.


Documentation
-------------
No formal documentation other than comments, source code, and the example Frogger application.

The API is mostly self-explanatory, though inspection of the example Frogger app will yield the most rapid understanding.


Known Issues
-----------------------

GoLang
* Datatypes
  excessive casting and use of int64 creating a large memory footprint for display objects.
* Error Handling
  The SDL layer surfaces many errors which GAS mostly discards. This is one of the reasons gas is not production ready.
* Public API
  Little thought into what fields should allow public access
* Recycling
  This systems should recycle dobs to avoid garbage collection. Currently no pooling of textures, dobs, etc.
* z-layering
  Controlled by OrderedMaps, which I have not benchmarked, tested thoroughly.









## License
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.



