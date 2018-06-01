Go Nano
=======

An implementation of the [Nano](https://nano.org/) protocol written from scratch in Go (golang).

About the Project
-----------------

A crypto currency has to be resilient to survive, and the network is only as resilient as the weakest link. With only one implementation of the protocol, any bugs that are found affect the entire network. The aim of this project is to create an alternative implementation that is 100% compatible with the reference implementation to create a more robust network.

Additionally, there is no reference specification for the Nano protocol, only a high level overview. I've had to learn the protocol from reading the source-code. I'm hoping a second implementation will be useful for others to learn the protocol.

Components
----------

Eventually the project will contain the following components:

 * [GoNano](https://github.com/frankh/nano)
    > A support library containing common functions, e.g. block validation, hashing, proof of work, etc
 * [Nano Vanity](https://github.com/frankh/nano-vanity)
    > A tool to generate vanity addresses (See https://en.bitcoin.it/wiki/Vanitygen)
 * [GoNano Node](#) - Coming Soon...
    > A full node implementation compatible with the official Nano wallet, but with faster initial sync times out of the box.
 * [GoNano Wallet](#) - Coming Soon...
    > A GUI Wallet that can either run as a node or as a light wallet.

Milestones
----------

  * ~Vanity Address Generator~
    > A simple project to get the basic public-key cryptography functions working and tested.
    - Done! ([Nano Vanity](https://github.com/frankh/nano-vanity))
  * GoNano Node
    * A basic node that can validate and store blocks sent to it
        * ~Data structures~
        * ~Database~
        * ~Proof of work~
        * ~Cryptographic functions~
        * ~Basic wallet functions~
        * Networking
            * ~Receiving keepalives and blocks~
                - ~keepalives~
                - ~publish~
                - ~confirm_ack~
            * Sending keepalives
    * Add broadcasting and discovery
    * Add RPC interface
    * Add voting
    * Add compatibility with existing Nano Nodes
    * Add spam defence and blacklisting of bad nodes
    * Add complete testing harness
    * Add fast syncing
  * GoNano Wallet
    * Basic UI, creating/sending/receiving transactions
    * Add seed restore, account generation, changing representatives
    * Add bundled node and light wallet/node selection option
    * UI Polish and distributables

Contributing
============

Any pull requests would be welcome!

I haven't been using Go for very long so any style/organisation fixes would be greatly appreciated.

Feel free to donate some Nano to nano_1frankh36p3e4cy4xrtj79d5rmcgce9wh4zke366gik19gifb5kxcnoju3y5 to help keep me motivated :beers:

