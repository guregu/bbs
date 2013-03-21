bbs protocol
=========

This is the start a generic protocol for message boards (internet forums, imageboards...). 
The idea is that we create a beautiful JSON protocol that will let us, for example, browse our favorite not-so-mobile-compliant message boards with a generic mobile app. 

I hope for this project to be:

  - Simple. Like IRC.
  - Flexible. Hopefully this will be able to encompass phpBB-like forums, reddit-like sites, 4chan-like imageboards...

 

Everything is still in the early stages. I wrote this all in a furious all nighter yesterday. The spec is still under construction but check out proto.go for an idea. bbs.go is a kind of generic router that you could write your own message board from. There is also a tiny command line client for testing things.

Right now my focus is on writing relays that translate data (via APIs or scraping), which are located here: https://github.com/tiko-chan/relay

Once we figure out how to deal with real world data, we can cook up our own message board software and plugins.


License
-

WTFPL? :)
