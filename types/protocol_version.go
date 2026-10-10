package types

type ProtocolId = uint16

type ProtocolVersion struct {
	ID    ProtocolId
	Names []string
}

var ProtocolVersions = struct {
	ZERO              ProtocolVersion
	MINECRAFT_1_7_2   ProtocolVersion
	MINECRAFT_1_7_6   ProtocolVersion
	MINECRAFT_1_8     ProtocolVersion
	MINECRAFT_1_9     ProtocolVersion
	MINECRAFT_1_9_1   ProtocolVersion
	MINECRAFT_1_9_2   ProtocolVersion
	MINECRAFT_1_9_3   ProtocolVersion
	MINECRAFT_1_10    ProtocolVersion
	MINECRAFT_1_11    ProtocolVersion
	MINECRAFT_1_11_1  ProtocolVersion
	MINECRAFT_1_12    ProtocolVersion
	MINECRAFT_1_12_1  ProtocolVersion
	MINECRAFT_1_12_2  ProtocolVersion
	MINECRAFT_1_13    ProtocolVersion
	MINECRAFT_1_13_1  ProtocolVersion
	MINECRAFT_1_13_2  ProtocolVersion
	MINECRAFT_1_14    ProtocolVersion
	MINECRAFT_1_14_1  ProtocolVersion
	MINECRAFT_1_14_2  ProtocolVersion
	MINECRAFT_1_14_3  ProtocolVersion
	MINECRAFT_1_14_4  ProtocolVersion
	MINECRAFT_1_15    ProtocolVersion
	MINECRAFT_1_15_1  ProtocolVersion
	MINECRAFT_1_15_2  ProtocolVersion
	MINECRAFT_1_16    ProtocolVersion
	MINECRAFT_1_16_1  ProtocolVersion
	MINECRAFT_1_16_2  ProtocolVersion
	MINECRAFT_1_16_3  ProtocolVersion
	MINECRAFT_1_16_4  ProtocolVersion
	MINECRAFT_1_17    ProtocolVersion
	MINECRAFT_1_17_1  ProtocolVersion
	MINECRAFT_1_18    ProtocolVersion
	MINECRAFT_1_18_2  ProtocolVersion
	MINECRAFT_1_19    ProtocolVersion
	MINECRAFT_1_19_1  ProtocolVersion
	MINECRAFT_1_19_3  ProtocolVersion
	MINECRAFT_1_19_4  ProtocolVersion
	MINECRAFT_1_20    ProtocolVersion
	MINECRAFT_1_20_2  ProtocolVersion
	MINECRAFT_1_20_3  ProtocolVersion
	MINECRAFT_1_20_5  ProtocolVersion
	MINECRAFT_1_21    ProtocolVersion
	MINECRAFT_1_21_2  ProtocolVersion
	MINECRAFT_1_21_4  ProtocolVersion
	MINECRAFT_1_21_5  ProtocolVersion
	MINECRAFT_1_21_6  ProtocolVersion
	MINECRAFT_1_21_7  ProtocolVersion
	MINECRAFT_1_21_9  ProtocolVersion
	MINECRAFT_1_21_11 ProtocolVersion
	MINECRAFT_26_1    ProtocolVersion
	MINECRAFT_26_2    ProtocolVersion
	MINECRAFT_26_3    ProtocolVersion
}{
	ZERO: ProtocolVersion{ID: 0, Names: []string{}},

	// The four releases from 1.7.2 to 1.7.5 share a protocol, 4, so a client
	// on any of them is a client on this version, the oldest this server
	// speaks. 1.7.6 moved to 5 and changed nothing in the numbering: every
	// id this server speaks is the same on the two, in every phase and both
	// directions, read off the 1.7.2 jar's registrations beside 1.7.6's, and
	// neither has an id for the set compression, which 1.8 added: see
	// HasCompression. The two jars were compared class by class with the
	// names taken out, and of the ninety-nine packets each registers, five
	// differ. 1.7.6 is where the uuid became a java.util.UUID: its login
	// success and its spawn player parse the text 1.7.2 keeps as text, and
	// its login start keeps one where 1.7.2 keeps none, which is the same
	// bytes on the wire for all three. The spawn player is the one packet
	// this server sends that differs: 1.7.6 reads the skin properties behind
	// the name, each a name, a value and a signature behind a count, and
	// 1.7.2 reads the position straight after the name, having no skin to
	// read. The set slot's window id is a short where 1.7.6 reads a byte,
	// and the plugin message's payload is capped at a short where 1.7.6 caps
	// it at a megabyte, and this server sends neither. The blocks are
	// 1.7.6's to the last number, read off the two jars' registrations side
	// by side: see package gamedata, and the 1.7.6 step's transformer.
	MINECRAFT_1_7_2: ProtocolVersion{ID: 4, Names: []string{"1.7.2", "1.7.3", "1.7.4", "1.7.5"}},

	// The five releases from 1.7.6 to 1.7.10 share a protocol, 5, so a client
	// on any of them is a client on this version: 1.7.2 sits on 4 below it,
	// and the 1.7.6 step carries everything down to it. 1.7.6 is where the
	// uuid went into the login success and the spawn player as a
	// java.util.UUID, and the skin properties behind the spawn player's
	// name, which is what 1.7.2 lacks. 1.8 moved to 47
	// and kept the play phase's numbering: every play id this server speaks
	// is the same on the two, read off the 1.7.6 jar's registrations beside
	// 1.8's, and so are the handshake, the status and the login phases, but
	// for the set compression packet, which 1.8 added and 1.7.6 has no id
	// for: a connection on 1.7.6 is never compressed. See HasCompression. Of
	// the packets laid out alike in the two, the animate, the login success,
	// the two disconnects and the status are; and everything else this
	// server sends differs. The keep alive carries its id as an int in both
	// directions where 1.8 reads a var int. The encryption request and its
	// response count their byte arrays with a short where 1.8 counts with a
	// var int. The play login ends at the level type, without 1.8's reduced
	// debug flag. The spawn position is three ints where 1.8 packs a long.
	// The player position ends with an on-ground flag where 1.8 ends with
	// relative flags, and carries the player's eye rather than its feet:
	// 1.7.6 keeps a player a block and five eighths above where 1.8 keeps
	// one, so its server adds 1.62 to the height it sends, and its client
	// sends two heights back, the feet and the eye, where 1.8's sends one.
	// The spawn player carries the uuid as text, and the name and the skin
	// properties behind it, where 1.8 carries the uuid alone and has the
	// rest from the player list, and crashes on a spawn player whose
	// metadata holds nothing; and the player list is a name, whether the
	// player is on and a ping, one player to a packet, where 1.8 lists
	// actions over entries keyed by uuid. The teleport, the head rotation
	// and the entity metadata name their entity with an int where 1.8 names
	// it with a var int, the teleport without 1.8's on-ground flag behind,
	// and the entity removal counts its ints with a byte. The chunk is the
	// map chunk bulk 1.8 reads, but with the chunks' bytes deflated and laid
	// out as every section's block ids, a byte each, then their variants,
	// a nibble each, then the light, then the biomes, where 1.8 reads a
	// short to a block, and with a second mask for blocks past a byte,
	// which no block of 1.7.6's is. Of what this server reads, the swing
	// names an entity and an animation where 1.8 sends nothing, the player
	// command numbers its actions from one and counts them in a byte, and
	// the player input says whether to jump and whether to unmount in two
	// flags of their own. The blocks are 1.8's less the twenty-seven blocks
	// and the seven variants 1.8 added, under the same numbers: see package
	// gamedata, and the 1.8 step's transformers.
	MINECRAFT_1_7_6: ProtocolVersion{ID: 5, Names: []string{"1.7.6", "1.7.7", "1.7.8", "1.7.9", "1.7.10"}},

	// The ten releases of the 1.8 cycle share a protocol, 47, so a client on
	// any of them is a client on this version: 1.7.6 sits on 5 below it, and
	// the 1.8 step carries everything down to it. 1.9 moved to 107 and rebuilt the play phase: it registers its
	// packets afresh in both directions, so no play id is 1.8's, which the
	// id tables say, and the handshake, the status and the login phases are
	// numbered and laid out alike. The 1.8 jar was read packet by packet, by
	// the calls each makes on the byte buffer, beside 1.9's. Of the play
	// packets this server sends, the keep alive, the login, the animate, the
	// player list, the entity removal, the head rotation and the spawn
	// position are laid out alike; and four are not, plus the chunk. The
	// player position lacks the teleport id 1.9 put on its end, and a 1.8
	// client acknowledges no teleport. The spawn player and the teleport
	// hold a position as three ints of thirty-seconds of a block where 1.9
	// holds three doubles, and the spawn player carries the held item as a
	// short in front of the metadata as well. The entity metadata is the
	// older form: each entry a byte that packs the kind in its top three
	// bits and the index in the low five, then the value, and the run ends
	// with 0x7F where 1.9 ends with 0xFF and names the kind in a var int of
	// its own. And the chunk knows no palette: a section is its blocks as
	// little-endian shorts, a block's number shifted up four bits with its
	// variant in the low four, the sections' block light behind all of them
	// and their sky light behind that, with the mask a short in front; and
	// a whole chunk with no section in it is one the client unloads. Of
	// what this server reads, the swing carries no hand, and the player
	// command's actions stop at opening the inventory, which sits one lower
	// than on 1.9. The blocks are 1.9's less the eighteen 1.9 added, under
	// the same numbers: see package gamedata, and the 1.9 step's
	// transformers.
	MINECRAFT_1_8: ProtocolVersion{ID: 47, Names: []string{"1.8", "1.8.1", "1.8.2", "1.8.3", "1.8.4", "1.8.5", "1.8.6", "1.8.7", "1.8.8", "1.8.9"}},

	// 1.9 has 107, its own: 1.8 sits on 47 below it, and the 1.9 step carries
	// everything down to it. 1.9.1
	// moved to 108 and shifted its jar's obfuscated names from the blocks'
	// tile entities on, so the two were compared class by class with the
	// names taken out. Every phase registers the same packets in the same
	// order in both, and every registered packet is the same class with the
	// names taken out but one, the play login: 1.9 reads its dimension as a
	// byte where 1.9.1 reads an int, and nothing else in it moved. The
	// packet buffer differs in the caps 1.9.1 put on its array readers,
	// which read the same bytes; the data watcher reads a serializer's id as
	// a byte where 1.9.1 reads a var int, which are the same byte for the
	// ids there are; the chunk, the spawn player, the entity metadata, the
	// entity's keys, the section, the block state container, the palettes,
	// the bit array and the nibble arrays are the same classes with the
	// names taken out; and the block state registry dumped slot by slot off
	// the 1.9 jar is 1.9.1's to the last state. So 107 is 108 under every
	// id, and the 1.9.1 step turns the login's dimension into a byte and
	// carries everything else as it is.
	MINECRAFT_1_9: ProtocolVersion{ID: 107, Names: []string{"1.9"}},

	// 1.9.1 has 108, its own. 1.9.2
	// moved to 109 without shifting its jar's obfuscated names: the two jars
	// hold the same classes under the same names, and eighteen differ, every
	// one for the version's number, its name or the data version its worlds
	// are stamped with, but the packet buffer, whose reader of a var int
	// array with no cap of its own caps the count at a quarter of the bytes
	// left rather than at all of them -- a cap one packet reads under, the set
	// passengers, which this server never sends. Every phase registers the
	// same packets in the same order in both, and every registered packet is
	// the same class with the names taken out; the block state registry
	// dumped slot by slot off the 1.9.1 jar is 1.9.2's to the last state. So
	// 108 is 109 under every id, and the 1.9.2 step carries everything as it
	// is.
	MINECRAFT_1_9_1: ProtocolVersion{ID: 108, Names: []string{"1.9.1"}},

	// 1.9.2 has 109, its own. 1.9.3
	// moved to 110 and shifted its jar's obfuscated names, so the two were
	// compared class by class with the names taken out. Every phase registers
	// the same packets in the same order in both but one: 1.9.2's play phase
	// registers seventy-seven clientbound packets to 1.9.3's seventy-six, the
	// one more being the update sign at 0x46, which 1.9.3 dropped, so every
	// clientbound play id from 0x47 up sits one higher on 1.9.2 -- the
	// teleport this server sends at 0x4A rather than 0x49 -- and nothing
	// below it moved. Of the packets both register, two differ as classes:
	// the chunk, which 1.9.3 gave a list of block entities behind its
	// sections -- a count and that many compounds -- where 1.9.2's ends at
	// the sections, and the sound effect, which scales its position by a
	// constant that changed, not on the wire. The spawn player, the entity
	// metadata, the data watcher and its serializers, the packet buffer, the
	// section, the block state container, the palettes, the bit array and the
	// nibble arrays are the same classes with the names taken out; the block
	// class differs only in how it loads a constant; the entity defines the
	// same five keys in the same order; and the block state registry dumped
	// slot by slot off the 1.9.2 jar is 1.9.3's to the last state. So 109 is
	// 110 under ids one higher from 0x47 up: the 1.9.3 step cuts the block
	// entities off the chunk, and the keep alive, the tags that are not
	// there, the login that cannot be asked and 1.9.3's block numbers are
	// carried as they are: see package gamedata.
	MINECRAFT_1_9_2: ProtocolVersion{ID: 109, Names: []string{"1.9.2"}},

	// 1.9.3 has 110, shared with 1.9.4. 1.10 moved to 210 over the frostburn
	// update, and the two jars were
	// compared class by class with the names taken out: every phase registers
	// the same packets in the same order in both -- one handshake, thirty
	// serverbound and seventy-six clientbound play packets, two and two
	// status, two and four login -- so not an id moved, and of the eight
	// registered packets whose class differs, none is one this server speaks:
	// the resource pack status lost its hash, the two sound effects' pitch
	// went from a byte to a float, the plugin message checks its payload for
	// null, the spawn mob and the entity velocity clamp their velocity by a
	// constant rather than a variable, the destroy entities writes its ids
	// through an index rather than an iterator, and the tab complete's field
	// turned final. The chunk packet, the spawn player, the entity metadata,
	// the data watcher and its serializers are the same classes with the
	// names taken out, as are the palettes, the bit array and the nibble
	// arrays; the block state container zeroes its bits before setting them
	// and the section's fields turned final, neither on the wire; the packet
	// buffer writes its arrays through an index rather than an iterator; and
	// the entity's flags are the first thing it defines in both, 1.10 adding
	// a sixth boolean -- no gravity -- behind the five 1.9.3 defines. So 110
	// is 210 under the same ids: 1.12.2's chunk, its keep alive carried
	// across the 1.12.2 step, no tags, a login that cannot be asked, and
	// 1.10's block numbers but for the five blocks 1.10 added -- the magma
	// block, the nether wart block, the red nether bricks, the bone block and
	// the structure void -- which 1.9.3 has no number for: see package
	// gamedata.
	MINECRAFT_1_9_3: ProtocolVersion{ID: 110, Names: []string{"1.9.3", "1.9.4"}},

	// 1.10 has 210, shared with 1.10.1 and 1.10.2. 1.11 moved to 315 over
	// the exploration update, and the
	// two jars were compared class by class with the names taken out: every
	// phase registers the same packets in the same order in both -- one
	// handshake, thirty serverbound and seventy-six clientbound play packets,
	// two and two status, two and four login -- so not an id moved, and of the
	// fourteen registered packets whose class differs, three are ones this
	// server speaks, and each differs outside what it writes: the spawn player
	// caches its metadata list on 1.10 and reads it fresh on 1.11, the chunk
	// works out whether its world has a sky one way and the other, and the
	// entity metadata clears its dirty flag on 1.11 and not on 1.10. The
	// chunk's sections, their palettes, the bit array and the nibble arrays
	// are the same classes with the names taken out, as are the metadata
	// serializers and their order, in which the byte is the first registered,
	// and the entity's flags are the first thing it defines in both. The
	// packet buffer differs in reading an item stack, which 1.11 made empty
	// rather than null, and in the wording of an error. So 210 is 315 under
	// the same ids: 1.12.2's chunk, its keep alive carried across the 1.12.2
	// step, no tags, a login that cannot be asked, and 1.11.1's block numbers
	// but for the seventeen blocks 1.11 added -- the observer and the sixteen
	// shulker boxes -- which 1.10 has no number for: see package gamedata.
	MINECRAFT_1_10: ProtocolVersion{ID: 210, Names: []string{"1.10", "1.10.1", "1.10.2"}},

	// 1.11 has 315 to itself. 1.11.1
	// moved to 316 over nothing on the wire: the two jars register the same
	// packets in the same order in every phase and direction -- one
	// handshake, thirty serverbound and seventy-six clientbound play packets,
	// two and two status, two and four login -- and each is the same class
	// with the names taken out, as are the connection state that lists them,
	// the packet buffer and the bootstrap. 1.11.1 was a release of fixes --
	// the iron nugget, the sweeping edge, a shield's cooldown -- and the
	// classes that differ are theirs; it added no block, and its registry
	// numbers every state as 1.11's does, slot for slot, read side by side
	// off the two jars. So 315 is 316 under the same ids: see package
	// gamedata for the numbers it shares.
	MINECRAFT_1_11: ProtocolVersion{ID: 315, Names: []string{"1.11"}},

	// 1.11.1 has 316, shared with 1.11.2.
	// 1.12 moved to 335 over the recipe book and the advancements:
	// it added the prepare crafting grid, the crafting book data and the
	// advancement tab to the serverbound play phase, the unlock recipes, the
	// advancement tab selection and the advancements to the clientbound one,
	// and put the base of the player moves and of the entity moves in front
	// of the three moves of each, where 1.11.1 has it behind them, so the
	// play ids sit apart on the two versions from each of those on, and
	// nothing else: the handshake, the status and the login phases are
	// numbered alike. 1.12 was a release of content,
	// and the jars differ in most of their classes; every packet this server
	// sends or reads was compared class by class with the names taken out,
	// and each is the same but for a log call's shape or a Guava helper's
	// name, as are the packet buffer, the chunk, its sections and their
	// palettes, the entity metadata and its serializers, in which the byte
	// is the first registered in both. So 316 is 335 under other play ids:
	// 1.12.2's chunk, its keep alive carried across the 1.12.2 step, no
	// tags, a login that cannot be asked, and 1.12.2's block numbers and
	// variants but for the eighteen blocks 1.12 added, which 1.11.1 has no
	// number for: see package gamedata.
	MINECRAFT_1_11_1: ProtocolVersion{ID: 316, Names: []string{"1.11.1", "1.11.2"}},

	// 1.12 has 335 to itself. 1.12.1
	// moved to 338 over the recipe book: it took out the prepare crafting
	// grid, the second packet of the serverbound play phase, and added the
	// craft recipe request further down it and the craft recipe response to
	// the clientbound one, so the serverbound play ids between those two, and
	// the clientbound ones from the response on, sit one apart on the two
	// versions. The jars were compared class by class
	// with the names taken out, 1.12.1's having shifted: of the 3,279 classes
	// 1.12 has, 38 differ, and two of them are packets -- the prepare
	// crafting grid, which 1.12.1 has not, and the handshake, for the number
	// it carries -- beside the play phase's registration, which lists them.
	// Every packet this server sends or reads is the same class to the byte,
	// as are the packet buffer and the block class, so 335 is 338 under other
	// play ids: 1.12.2's block numbers and variants, its chunk, its keep
	// alive carried across the 1.12.2 step, no tags, and a login that cannot
	// be asked. The rest of the 38 are the recipe book and its screens, the
	// version strings and the data version, 1139 against 1241.
	MINECRAFT_1_12: ProtocolVersion{ID: 335, Names: []string{"1.12"}},

	// 1.12.1 has 338 to itself. 1.12.2
	// is a release of fixes, and its jar differs from 1.12.1's in fifty
	// classes, of which three are packets: the two keep alives, whose id is
	// a var int on 1.12.1 and a long on 1.12.2, in both directions, and the
	// handshake, for the number it carries. Every other packet this server
	// sends or reads is the same class to the byte, every phase registers the
	// same packets in the same order, and the block class is the same too,
	// so 338 draws 1.12.2's block numbers and variants. The keep alive is
	// carried across by the 1.12.2 step's transformers, which is why the id
	// this server picks fits a var int: see client.SendKeepAlive.
	MINECRAFT_1_12_1: ProtocolVersion{ID: 338, Names: []string{"1.12.1"}},

	// 1.12.2 has 340 to itself. It is
	// from before the flattening: a client on it knows a block by a number
	// and a four-bit variant beside it rather than by a state of its own, so
	// a chunk goes out to it in block ids that 1.13's own data fixers reach
	// from the other side (see package gamedata), packed thirteen bits to an
	// entry when a section is too varied for a palette. Its chunk is 1.13's
	// otherwise, but for a count of nothing in front of the ids a section
	// names directly, and for the biomes, which are a byte each. It has no
	// tags packet at all, and no login plugin message either, so a proxy
	// can forward a login to it only in the handshake: see
	// HasLoginPluginMessages. Everything else this server says or reads is
	// laid out as on 1.13 and numbered otherwise: see the 1.13 step's
	// transformers.
	MINECRAFT_1_12_2: ProtocolVersion{ID: 340, Names: []string{"1.12.2"}},

	// 1.13 has 393 to itself. 1.13.1
	// is a release of fixes that reach nothing this server says or reads, and
	// its jar kept 1.13's names for the classes on the wire, so the two were
	// compared class by class with the names taken out: every phase registers
	// the same packets in the same order, 142 of them under the same names,
	// and none of the packets that read or write differently is one this
	// server sends or reads. 1.13.1 moved the game's registries behind one
	// interface, which the statistics, the block event, the particles and the
	// effects reach their ids through; the play login and the respawn keep
	// their dimension as a type and write its number as before; the command
	// suggestion caps its text; and the book edit took on a hand. The byte
	// buffer, the metadata and its serializers but the particle's, the order
	// the entity, the living entity and the player define theirs in, the
	// palettes, the bit storage, the chunk section, the chunk packet and the
	// tags packet are the same code. What differs is data: 1.13.1 added the
	// five dead coral plants, gave the corals and the conduit a waterlogged
	// property and the TNT an unstable one, and added the coral plants and
	// the underwater bonemeals tags, so 393 has a block state table and a set
	// of tags of its own. Everything else said of 1.13.1 below is as true of
	// 393.
	MINECRAFT_1_13: ProtocolVersion{ID: 393, Names: []string{"1.13"}},

	// 1.13.1 has 401 to itself. 1.13.2
	// is a release of fixes, and of one change on the wire that reaches
	// nothing this server says or reads: an item stack, which 1.13.1 writes as
	// a short item id, -1 for none, and 1.13.2 as a flag for whether there is
	// one and a varint item id -- and no packet this server sends or reads
	// carries an item. Its jar shifted the obfuscated names past the fish, a
	// class of which it split in two, so the two were compared class by class
	// with the names taken out: every phase registers the same packets in the
	// same order, 142 of them, under the same names, and every one reads and
	// writes the same fields in the same order. Of the 97 classes that differ
	// once the names are out, the byte buffer is the one on the wire, in its
	// item stack alone. The chunk loads its block entities from a save
	// another way, and reads a chunk packet's data as before; a chunk being
	// generated marks fewer blocks for post-processing; the rest are the fish and their
	// buckets, the data fixers, the data version and the version strings, the
	// structures, and the client's screens and renderers. The two jars'
	// blocks, commands and items reports are the same bytes, and every file
	// that is not a class is the same, the tags among them. So everything said
	// of 1.13.2 below is as true of 401.
	MINECRAFT_1_13_1: ProtocolVersion{ID: 401, Names: []string{"1.13.1"}},

	// 1.13.2 has 404 to itself. It is
	// from before the light left the chunk packet: a client on it reads a
	// section's light behind its blocks, in the chunk itself, and has no
	// packet of its own for either the light or the chunk its view is
	// centred on. It is also from before the heightmaps went on the wire,
	// before an entity had a pose, before the play login lost the
	// difficulty and took on the view distance, before the entity types had
	// tags, and before a block position kept its height in its lowest bits:
	// see the 1.14 step's transformers, and package world.
	MINECRAFT_1_13_2: ProtocolVersion{ID: 404, Names: []string{"1.13.2"}},

	// 1.14 has 477 to itself: 1.13.2 sits on 404 below it. 1.14.1 is
	// a release of fixes that reach nothing this server says or reads. Its jar
	// shifted the obfuscated names, so the two were compared class by class
	// with the names taken out: every phase registers the same packets in the
	// same order, 151 of them, and the bodies of two differ. The custom payload
	// gained a channel for debugging raids, which no server sends a player,
	// and the explosion floors its centre where 1.14's truncated it, writing
	// and reading the same fields. The byte buffer, the metadata and its
	// serializers, the poses, the palettes, the chunk section, the heightmap,
	// the chunk and light packets, the login and the tags packet are the same
	// code, and so are the client's chunk cache, its level, the local player,
	// the living entity and the player. Of what differs behind them, the
	// entity builds a block position out of its doubles rather than flooring
	// them itself, its metadata defined in the same order; the chunk moved the
	// unpacking of its tick lists into a method of its own, which a server
	// runs; the tag collection loads its files another way and writes them as
	// before; and the packet listener reads the debug payloads 1.14.1 changed.
	// The two jars' blocks and commands reports are the same bytes, their tags
	// the same files with the same contents, and their registries differ by
	// one memory module a villager's brain keeps on the server. So everything
	// said of 1.14.1 below is as true of 477.
	MINECRAFT_1_14: ProtocolVersion{ID: 477, Names: []string{"1.14"}},

	// 1.14.1 has 480 to itself. 1.14.2
	// is a release of fixes that reach nothing this server says or reads, and
	// its jar kept 1.14.1's obfuscated names, so the two compare entry by
	// entry: 209 classes differ and one is new, a part of the chunk status.
	// None of them is a packet -- all 151 are the same bytes, and so is the
	// class that registers them -- nor the byte buffer, the metadata and its
	// serializers, the tags, the bit storage, the palettes, the chunk section
	// or the heightmap. Of what stands behind them, the chunk gained a way to
	// hand out a block entity's tag and changed nothing of how it reads a
	// packet; the entity, the living entity and the player define their
	// metadata in the same order; the client's chunk cache gained three ways to
	// ask whether it holds a chunk; the local player changed how it takes up
	// sprinting and nothing of what it sends; and 1.14.2's packet listener adds
	// a chunk's entities only when the chunk came whole, where 1.14.1's adds
	// them for a part of one as well. The chat components took on a depth to stop
	// resolving at, on the server's side, and are written as before. The two
	// jars' blocks reports are the same bytes, and their tags the same files
	// by name: 1.14.2 put the cut sandstone slabs among the slabs, which is
	// the contents of a tag, and a tag is sent here by its name alone. So
	// everything said of 1.14.2 below is as true of 480.
	MINECRAFT_1_14_1: ProtocolVersion{ID: 480, Names: []string{"1.14.1"}},

	// 1.14.2 has 485 to itself. 1.14.3
	// is a release of fixes that reach nothing this server says or reads. Its
	// jar shifted the obfuscated names, so the two were compared class by class
	// with the names taken out: every phase registers the same packets in the
	// same order, and of the play phase's 138 the ones that differ in what
	// they write are seven this server never sends -- the add entity, the add
	// mob, the add painting, the merchant offers, the entity motion, the
	// passengers and the recipes. The byte buffer, the login, the chunk, the
	// light, the tags and the entity data packets, the metadata's serializers,
	// the bit storage and the palettes are the same classes, and the order a
	// player's metadata is defined in is the same. The tags are written and
	// read as the same run of numbers: 1.14.3 builds the map it reads aside and
	// swaps it in whole, where 1.14.2 fills the one in use. The entity data
	// gained a check in 1.14.3 that a value is of the type its field was
	// defined with, which 1.14.2 lacks and nothing sent here would fail. The
	// two jars' tags are the same files and their blocks reports the same
	// bytes. So everything said of 1.14.3 below is as true of 485.
	MINECRAFT_1_14_2: ProtocolVersion{ID: 485, Names: []string{"1.14.2"}},

	// 1.14.3 has 490 to itself. 1.14.4
	// is a release of fixes that reach nothing this server says or reads. Its
	// jar was obfuscated afresh, so the two were compared class by class with
	// the names taken out: the play phase registers 1.14.3's packets in
	// 1.14.3's order and then one more, the block break acknowledgement, which
	// 1.14.4 added as the last the server sends and this server never does. Of
	// the packets the two share, three differ -- the add player and the player
	// info by the name of a method they call, the player action by copying the
	// position it is given -- and none in what it reads or writes. The byte
	// buffer, the entity metadata's serializers and the order a player's
	// metadata is defined in are the same; the bit storage and the palettes
	// gained a way of counting what they hold, not of writing it. The two jars'
	// tags are the same files and their blocks reports the same bytes. So
	// everything said of 1.14.4 below is as true of 490.
	MINECRAFT_1_14_3: ProtocolVersion{ID: 490, Names: []string{"1.14.3"}},

	// 1.14.4 has 498 to itself. It is
	// from before the biomes of a chunk took on a height: a client on it holds
	// one for every column and reads them off the end of the chunk's sections,
	// where 1.15 reads one for every four blocks each way out of the packet
	// itself. It is also from before the play login held the seed and said
	// whether a death is followed by its screen, and from while the add player
	// packet still carried the player's entity metadata: see the 1.15 step's
	// transformers.
	MINECRAFT_1_14_4: ProtocolVersion{ID: 498, Names: []string{"1.14.4"}},

	// 1.15 has 573 to itself: the numbers between it and 498 went to its
	// snapshots. 1.15.1 is
	// a release of fixes that reach nothing a server says: its jar is 1.15's
	// class for class but for fifty-one of them, none a packet or a class that
	// reads or writes one, and its tags and its data generator's reports are
	// the same files. The one of them a packet reaches is the chunk's biomes,
	// which 1.15 reads as 1.15.1 does but keeps a number it does not know as no
	// biome at all, where 1.15.1 puts a default in its place: this server sends
	// the plains, which both know. So everything said of 1.15.1 and 1.15.2
	// below is as true of 573.
	MINECRAFT_1_15: ProtocolVersion{ID: 573, Names: []string{"1.15"}},

	// 1.15.1 has 575 to itself. 1.15.2
	// is a release of fixes that reach nothing a server says: its jar holds
	// every packet and serialization class of 1.15.1's as it stands, registers
	// the same packets in the same order, and its tags and its data
	// generator's block and registry reports are the same files. So everything
	// said of 1.15.2 below is as true of 575.
	MINECRAFT_1_15_1: ProtocolVersion{ID: 575, Names: []string{"1.15.1"}},

	// 1.15.2 has 578 to itself. It is
	// from before the dimension types became data: a client on it knows every
	// dimension for itself, as it knows every biome, and is put into one by
	// number. It is also from before a packed entry stopped crossing from one
	// long into the next, which is how it reads a chunk's blocks and its
	// heightmaps, and from before the login success held the uuid as
	// anything but text: see the 1.16 step's transformers.
	MINECRAFT_1_15_2: ProtocolVersion{ID: 578, Names: []string{"1.15.2"}},

	// 1.16 has 735 to itself: the numbers between it and 578 went to its
	// snapshots. 1.16.1 is
	// a fix to the Realms screens: its jar is 1.16's, class for class, but for
	// the version it names and the Realms main and upload screens, and its
	// data is the same files. So everything said of 1.16.1 below is as true of
	// 735.
	MINECRAFT_1_16: ProtocolVersion{ID: 735, Names: []string{"1.16"}},

	// 1.16.1 has 736 to itself. It is from before the registries went into the
	// play login: a client on it knows every biome for itself, is sent the
	// dimension types alone, and is put into one of them by name. It is also
	// from before the chunk packet lost its second flag and its biomes took
	// on a count: see the 1.16.2 step's transformers.
	MINECRAFT_1_16_1: ProtocolVersion{ID: 736, Names: []string{"1.16.1"}},

	// 1.16.2 has 751 to itself: the numbers between it and 736 went to its
	// snapshots. 1.16.3 is a fix to how mobs find their way: its jar is
	// 1.16.2's, class for class, but for the version it names and five
	// classes of the mob and its pathfinding, and its data is the same files.
	// 752 was a release candidate's number on the way, which no release
	// speaks, so everything said of 1.16.3 and 1.16.4 below is as true of 751.
	MINECRAFT_1_16_2: ProtocolVersion{ID: 751, Names: []string{"1.16.2"}},

	// 1.16.3 has 753 to itself: 1.16.4 moved to 754 for its social
	// interactions screen, which is the client's own business, and changed
	// nothing this server sends or reads -- not an id, not a packet, not a
	// tag, not a block. Everything said of 1.16.4 below is as true of it.
	MINECRAFT_1_16_3: ProtocolVersion{ID: 753, Names: []string{"1.16.3"}},

	// 1.16.5 stayed on 1.16.4's protocol, so a client on either of them is a
	// client on this version. With 1.16.3 and 1.16.2 below it, it is from before the
	// world grew downwards: a client on it holds a world of sixteen sections
	// from zero up, whatever the dimension type says, and reads a chunk and
	// its light behind masks that are plain numbers: see the 1.17 step's
	// transformers.
	MINECRAFT_1_16_4: ProtocolVersion{ID: 754, Names: []string{"1.16.4", "1.16.5"}},

	// 1.17 has 755 to itself: 1.16.4 sits on 754 and 1.17.1 moved to 756 over
	// one packet, the one that takes an entity out of the world, which 1.17
	// reads one entity at a time. In everything else it is
	// 1.17.1: from before the configuration phase, from before the profile
	// key, and from before the light joined the chunk packet.
	MINECRAFT_1_17: ProtocolVersion{ID: 755, Names: []string{"1.17"}},

	// 1.17.1 has 756 to itself: 1.17 sits on 755 and 1.18 moved to 757. Like
	// the seven above it, it is from
	// before the configuration phase: see HasConfigurationPhase. It is from
	// before the profile key as well, so a client on it always encrypts the
	// encryption challenge: see MaySignEncryptionChallenge. It is also from
	// before the light joined the chunk packet, so a client on it reads the
	// light in a packet of its own: see package world.
	MINECRAFT_1_17_1: ProtocolVersion{ID: 756, Names: []string{"1.17.1"}},

	// 1.18.1 stayed on 1.18's protocol, so a client on either of them is a
	// client on this version. 1.17.1 sits on 756 below it. Like the six
	// above it, it is from before the configuration phase: see
	// HasConfigurationPhase. It is from before the profile key as well, so a
	// client on it always encrypts the encryption challenge: see
	// MaySignEncryptionChallenge.
	MINECRAFT_1_18: ProtocolVersion{ID: 757, Names: []string{"1.18", "1.18.1"}},

	// 1.18.2 has 758 to itself: 1.18 and 1.18.1 share 757 below it and 1.19
	// moved to 759. Like the five above it, it is from before the
	// configuration phase: see HasConfigurationPhase. It is from before the
	// profile key as well, so a client on it always encrypts the encryption
	// challenge: see MaySignEncryptionChallenge.
	MINECRAFT_1_18_2: ProtocolVersion{ID: 758, Names: []string{"1.18.2"}},

	// 1.19 has 759 to itself: 1.18.2 sits on 758 and 1.19.1 moved to 760.
	// Like the four above it, it is from before the configuration phase: see
	// HasConfigurationPhase. It is the first on which a client may sign the
	// encryption challenge rather than encrypt it, since 1.19 is where the
	// profile key appeared: see MaySignEncryptionChallenge.
	MINECRAFT_1_19: ProtocolVersion{ID: 759, Names: []string{"1.19"}},

	// 1.19.2 stayed on 1.19.1's protocol, so a client on either of them is a
	// client on this version. Like the three above it, it is from before the
	// configuration phase: see HasConfigurationPhase. It is also the last on
	// which a client may sign the encryption challenge rather than encrypt
	// it: see MaySignEncryptionChallenge.
	MINECRAFT_1_19_1: ProtocolVersion{ID: 760, Names: []string{"1.19.1", "1.19.2"}},

	// 1.19.3 has 761 to itself: 1.19.2 sits on 760 and 1.19.4 moved to 762.
	// It is from before the configuration phase as well.
	MINECRAFT_1_19_3: ProtocolVersion{ID: 761, Names: []string{"1.19.3"}},

	// 1.19.4 has 762 to itself: 1.19.3 sits on 761 and 1.20 moved to 763. It
	// is from before the configuration phase as well.
	MINECRAFT_1_19_4: ProtocolVersion{ID: 762, Names: []string{"1.19.4"}},

	// 1.20.1 stayed on 1.20's protocol, so a client on either of them is a
	// client on this version. It is the last before the configuration phase:
	// see HasConfigurationPhase.
	MINECRAFT_1_20: ProtocolVersion{ID: 763, Names: []string{"1.20", "1.20.1"}},

	// 1.20.2 has 764 to itself: 1.20 and 1.20.1 share 763 below it, and 1.20.3
	// moved to 765.
	MINECRAFT_1_20_2: ProtocolVersion{ID: 764, Names: []string{"1.20.2"}},

	// 1.20.4 stayed on 1.20.3's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_20_3: ProtocolVersion{ID: 765, Names: []string{"1.20.3", "1.20.4"}},

	// 1.20.6 stayed on 1.20.5's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_20_5: ProtocolVersion{ID: 766, Names: []string{"1.20.5", "1.20.6"}},

	// 1.21.1 stayed on 1.21's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21: ProtocolVersion{ID: 767, Names: []string{"1.21", "1.21.1"}},

	// 1.21.3 stayed on 1.21.2's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21_2: ProtocolVersion{ID: 768, Names: []string{"1.21.2", "1.21.3"}},

	// 1.21.4 has 769 to itself: 1.21.2 and 1.21.3 share 768 below it, and
	// 1.21.5 moved to 770.
	MINECRAFT_1_21_4: ProtocolVersion{ID: 769, Names: []string{"1.21.4"}},

	// 1.21.5 has 770 to itself: 1.21.4 sits on 769 and 1.21.6 moved to 771.
	MINECRAFT_1_21_5: ProtocolVersion{ID: 770, Names: []string{"1.21.5"}},

	// 1.21.6 has 771 to itself: 1.21.5 sits on 770 and 1.21.7 moved to 772.
	MINECRAFT_1_21_6: ProtocolVersion{ID: 771, Names: []string{"1.21.6"}},

	// 1.21.8 stayed on 1.21.7's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21_7: ProtocolVersion{ID: 772, Names: []string{"1.21.7", "1.21.8"}},

	// 1.21.10 stayed on 1.21.9's protocol, so a client on either of them is a
	// client on this version; 1.21.11 has 774 to itself.
	MINECRAFT_1_21_9:  ProtocolVersion{ID: 773, Names: []string{"1.21.9", "1.21.10"}},
	MINECRAFT_1_21_11: ProtocolVersion{ID: 774, Names: []string{"1.21.11"}},

	// The three releases of the 26.1 cycle share a protocol, so a client on any
	// of them is a client on this version.
	MINECRAFT_26_1: ProtocolVersion{ID: 775, Names: []string{"26.1", "26.1.1", "26.1.2"}},

	// 26.2 has 776 to itself: the 26.1 releases share 775 below it and 26.3
	// moved to 777.
	MINECRAFT_26_2: ProtocolVersion{ID: 776, Names: []string{"26.2"}},

	// 26.3 is the latest version, the one everything is implemented at: see
	// LatestProtocolVersion.
	MINECRAFT_26_3: ProtocolVersion{ID: 777, Names: []string{"26.3"}},
}

// SupportedProtocolVersions is every version a client may connect on, oldest
// first.
//
// The order is the one the packet transformers walk. A packet read from a
// client is carried up this list one version at a time until it reaches the
// latest, which is the only version anything is implemented at, and a packet
// written to a client is carried back down it. That is why the list has to hold
// every version in between rather than only the ends: a step is what a
// transformer is registered for.
//
// ZERO is not among them. It is what a connection speaks before its handshake
// says otherwise, which is not a version anything is transformed to or from.
var SupportedProtocolVersions = []ProtocolVersion{
	ProtocolVersions.MINECRAFT_1_7_2,
	ProtocolVersions.MINECRAFT_1_7_6,
	ProtocolVersions.MINECRAFT_1_8,
	ProtocolVersions.MINECRAFT_1_9,
	ProtocolVersions.MINECRAFT_1_9_1,
	ProtocolVersions.MINECRAFT_1_9_2,
	ProtocolVersions.MINECRAFT_1_9_3,
	ProtocolVersions.MINECRAFT_1_10,
	ProtocolVersions.MINECRAFT_1_11,
	ProtocolVersions.MINECRAFT_1_11_1,
	ProtocolVersions.MINECRAFT_1_12,
	ProtocolVersions.MINECRAFT_1_12_1,
	ProtocolVersions.MINECRAFT_1_12_2,
	ProtocolVersions.MINECRAFT_1_13,
	ProtocolVersions.MINECRAFT_1_13_1,
	ProtocolVersions.MINECRAFT_1_13_2,
	ProtocolVersions.MINECRAFT_1_14,
	ProtocolVersions.MINECRAFT_1_14_1,
	ProtocolVersions.MINECRAFT_1_14_2,
	ProtocolVersions.MINECRAFT_1_14_3,
	ProtocolVersions.MINECRAFT_1_14_4,
	ProtocolVersions.MINECRAFT_1_15,
	ProtocolVersions.MINECRAFT_1_15_1,
	ProtocolVersions.MINECRAFT_1_15_2,
	ProtocolVersions.MINECRAFT_1_16,
	ProtocolVersions.MINECRAFT_1_16_1,
	ProtocolVersions.MINECRAFT_1_16_2,
	ProtocolVersions.MINECRAFT_1_16_3,
	ProtocolVersions.MINECRAFT_1_16_4,
	ProtocolVersions.MINECRAFT_1_17,
	ProtocolVersions.MINECRAFT_1_17_1,
	ProtocolVersions.MINECRAFT_1_18,
	ProtocolVersions.MINECRAFT_1_18_2,
	ProtocolVersions.MINECRAFT_1_19,
	ProtocolVersions.MINECRAFT_1_19_1,
	ProtocolVersions.MINECRAFT_1_19_3,
	ProtocolVersions.MINECRAFT_1_19_4,
	ProtocolVersions.MINECRAFT_1_20,
	ProtocolVersions.MINECRAFT_1_20_2,
	ProtocolVersions.MINECRAFT_1_20_3,
	ProtocolVersions.MINECRAFT_1_20_5,
	ProtocolVersions.MINECRAFT_1_21,
	ProtocolVersions.MINECRAFT_1_21_2,
	ProtocolVersions.MINECRAFT_1_21_4,
	ProtocolVersions.MINECRAFT_1_21_5,
	ProtocolVersions.MINECRAFT_1_21_6,
	ProtocolVersions.MINECRAFT_1_21_7,
	ProtocolVersions.MINECRAFT_1_21_9,
	ProtocolVersions.MINECRAFT_1_21_11,
	ProtocolVersions.MINECRAFT_26_1,
	ProtocolVersions.MINECRAFT_26_2,
	ProtocolVersions.MINECRAFT_26_3,
}

// LatestProtocolVersion is the version every packet is implemented at. Older
// versions are reached by transforming what this one produces, so adding one
// adds a table of ids and whatever transformers its differences need, and no
// second implementation of anything.
var LatestProtocolVersion = SupportedProtocolVersions[len(SupportedProtocolVersions)-1]

var protocolVersionsById = map[ProtocolId]ProtocolVersion{
	ProtocolVersions.ZERO.ID:              ProtocolVersions.ZERO,
	ProtocolVersions.MINECRAFT_1_7_2.ID:   ProtocolVersions.MINECRAFT_1_7_2,
	ProtocolVersions.MINECRAFT_1_7_6.ID:   ProtocolVersions.MINECRAFT_1_7_6,
	ProtocolVersions.MINECRAFT_1_8.ID:     ProtocolVersions.MINECRAFT_1_8,
	ProtocolVersions.MINECRAFT_1_9.ID:     ProtocolVersions.MINECRAFT_1_9,
	ProtocolVersions.MINECRAFT_1_9_1.ID:   ProtocolVersions.MINECRAFT_1_9_1,
	ProtocolVersions.MINECRAFT_1_9_2.ID:   ProtocolVersions.MINECRAFT_1_9_2,
	ProtocolVersions.MINECRAFT_1_9_3.ID:   ProtocolVersions.MINECRAFT_1_9_3,
	ProtocolVersions.MINECRAFT_1_10.ID:    ProtocolVersions.MINECRAFT_1_10,
	ProtocolVersions.MINECRAFT_1_11.ID:    ProtocolVersions.MINECRAFT_1_11,
	ProtocolVersions.MINECRAFT_1_11_1.ID:  ProtocolVersions.MINECRAFT_1_11_1,
	ProtocolVersions.MINECRAFT_1_12.ID:    ProtocolVersions.MINECRAFT_1_12,
	ProtocolVersions.MINECRAFT_1_12_1.ID:  ProtocolVersions.MINECRAFT_1_12_1,
	ProtocolVersions.MINECRAFT_1_12_2.ID:  ProtocolVersions.MINECRAFT_1_12_2,
	ProtocolVersions.MINECRAFT_1_13.ID:    ProtocolVersions.MINECRAFT_1_13,
	ProtocolVersions.MINECRAFT_1_13_1.ID:  ProtocolVersions.MINECRAFT_1_13_1,
	ProtocolVersions.MINECRAFT_1_13_2.ID:  ProtocolVersions.MINECRAFT_1_13_2,
	ProtocolVersions.MINECRAFT_1_14.ID:    ProtocolVersions.MINECRAFT_1_14,
	ProtocolVersions.MINECRAFT_1_14_1.ID:  ProtocolVersions.MINECRAFT_1_14_1,
	ProtocolVersions.MINECRAFT_1_14_2.ID:  ProtocolVersions.MINECRAFT_1_14_2,
	ProtocolVersions.MINECRAFT_1_14_3.ID:  ProtocolVersions.MINECRAFT_1_14_3,
	ProtocolVersions.MINECRAFT_1_14_4.ID:  ProtocolVersions.MINECRAFT_1_14_4,
	ProtocolVersions.MINECRAFT_1_15.ID:    ProtocolVersions.MINECRAFT_1_15,
	ProtocolVersions.MINECRAFT_1_15_1.ID:  ProtocolVersions.MINECRAFT_1_15_1,
	ProtocolVersions.MINECRAFT_1_15_2.ID:  ProtocolVersions.MINECRAFT_1_15_2,
	ProtocolVersions.MINECRAFT_1_16.ID:    ProtocolVersions.MINECRAFT_1_16,
	ProtocolVersions.MINECRAFT_1_16_1.ID:  ProtocolVersions.MINECRAFT_1_16_1,
	ProtocolVersions.MINECRAFT_1_16_2.ID:  ProtocolVersions.MINECRAFT_1_16_2,
	ProtocolVersions.MINECRAFT_1_16_3.ID:  ProtocolVersions.MINECRAFT_1_16_3,
	ProtocolVersions.MINECRAFT_1_16_4.ID:  ProtocolVersions.MINECRAFT_1_16_4,
	ProtocolVersions.MINECRAFT_1_17.ID:    ProtocolVersions.MINECRAFT_1_17,
	ProtocolVersions.MINECRAFT_1_17_1.ID:  ProtocolVersions.MINECRAFT_1_17_1,
	ProtocolVersions.MINECRAFT_1_18.ID:    ProtocolVersions.MINECRAFT_1_18,
	ProtocolVersions.MINECRAFT_1_18_2.ID:  ProtocolVersions.MINECRAFT_1_18_2,
	ProtocolVersions.MINECRAFT_1_19.ID:    ProtocolVersions.MINECRAFT_1_19,
	ProtocolVersions.MINECRAFT_1_19_1.ID:  ProtocolVersions.MINECRAFT_1_19_1,
	ProtocolVersions.MINECRAFT_1_19_3.ID:  ProtocolVersions.MINECRAFT_1_19_3,
	ProtocolVersions.MINECRAFT_1_19_4.ID:  ProtocolVersions.MINECRAFT_1_19_4,
	ProtocolVersions.MINECRAFT_1_20.ID:    ProtocolVersions.MINECRAFT_1_20,
	ProtocolVersions.MINECRAFT_1_20_2.ID:  ProtocolVersions.MINECRAFT_1_20_2,
	ProtocolVersions.MINECRAFT_1_20_3.ID:  ProtocolVersions.MINECRAFT_1_20_3,
	ProtocolVersions.MINECRAFT_1_20_5.ID:  ProtocolVersions.MINECRAFT_1_20_5,
	ProtocolVersions.MINECRAFT_1_21.ID:    ProtocolVersions.MINECRAFT_1_21,
	ProtocolVersions.MINECRAFT_1_21_2.ID:  ProtocolVersions.MINECRAFT_1_21_2,
	ProtocolVersions.MINECRAFT_1_21_4.ID:  ProtocolVersions.MINECRAFT_1_21_4,
	ProtocolVersions.MINECRAFT_1_21_5.ID:  ProtocolVersions.MINECRAFT_1_21_5,
	ProtocolVersions.MINECRAFT_1_21_6.ID:  ProtocolVersions.MINECRAFT_1_21_6,
	ProtocolVersions.MINECRAFT_1_21_7.ID:  ProtocolVersions.MINECRAFT_1_21_7,
	ProtocolVersions.MINECRAFT_1_21_9.ID:  ProtocolVersions.MINECRAFT_1_21_9,
	ProtocolVersions.MINECRAFT_1_21_11.ID: ProtocolVersions.MINECRAFT_1_21_11,
	ProtocolVersions.MINECRAFT_26_1.ID:    ProtocolVersions.MINECRAFT_26_1,
	ProtocolVersions.MINECRAFT_26_2.ID:    ProtocolVersions.MINECRAFT_26_2,
	ProtocolVersions.MINECRAFT_26_3.ID:    ProtocolVersions.MINECRAFT_26_3,
}

func GetProtocolVersionById(id ProtocolId) ProtocolVersion {
	if v, ok := protocolVersionsById[id]; ok {
		return v
	}

	return ProtocolVersions.ZERO
}

// protocolVersionIndex is where version sits in SupportedProtocolVersions, or
// -1 for a version that is not on the chain at all.
func protocolVersionIndex(version ProtocolVersion) int {
	for i, supported := range SupportedProtocolVersions {
		if supported.ID == version.ID {
			return i
		}
	}

	return -1
}

// IsSupportedProtocolVersion reports whether version is one the server speaks,
// which is to say one the transformers can carry a packet to and from.
func IsSupportedProtocolVersion(version ProtocolVersion) bool {
	return protocolVersionIndex(version) >= 0
}

// NextProtocolVersion returns the version one step newer than version. It
// reports false for the latest version, which has nothing above it, and for a
// version that is not on the chain.
func NextProtocolVersion(version ProtocolVersion) (ProtocolVersion, bool) {
	index := protocolVersionIndex(version)
	if index < 0 || index == len(SupportedProtocolVersions)-1 {
		return ProtocolVersions.ZERO, false
	}

	return SupportedProtocolVersions[index+1], true
}

// PreviousProtocolVersion returns the version one step older than version. It
// reports false for the oldest version and for a version that is not on the
// chain.
func PreviousProtocolVersion(version ProtocolVersion) (ProtocolVersion, bool) {
	index := protocolVersionIndex(version)
	if index <= 0 {
		return ProtocolVersions.ZERO, false
	}

	return SupportedProtocolVersions[index-1], true
}

// HasConfigurationPhase reports whether a client on this version passes
// through the configuration phase on its way from the login to the play
// phase. 1.20.2 is where the phase appeared. A client before it -- 1.20,
// 1.19.4, 1.19.3, 1.19.1, 1.19, 1.18.2, 1.18, 1.17.1, 1.17, 1.16.4, 1.16.3, 1.16.2, 1.16.1, 1.16, 1.15.2, 1.15.1, 1.15, 1.14.4, 1.14.3, 1.14.2, 1.14.1, 1.14, 1.13.2, 1.13.1, 1.13, 1.12.2, 1.12.1, 1.12, 1.11.1, 1.11, 1.10, 1.9.3, 1.9.2, 1.9.1, 1.9, 1.8, 1.7.6 and 1.7.2 -- is in play the
// moment its login succeeds, with nothing acknowledged in between, and what the phase carries
// from 1.20.2 on -- the registries and the tags -- reaches such a client
// through the play phase instead: the registries inside the play login
// packet itself, and the tags as a play packet right after it.
func (v ProtocolVersion) HasConfigurationPhase() bool {
	return v.ID >= ProtocolVersions.MINECRAFT_1_20_2.ID
}

// HasCompression reports whether a client on this version can be told a
// compression threshold, which is the set compression packet of the login
// phase. 1.8 is where the packet appeared, and with it the framing that puts
// a body's inflated size in front of it. A client before it -- 1.7.6 or
// 1.7.2 -- has no id for the packet and reads every frame as a length and a body, so a
// connection on it is never compressed: the packet is not sent, and nothing
// after it is framed as if it had been. See handlers.completeLogin.
func (v ProtocolVersion) HasCompression() bool {
	return v.ID >= ProtocolVersions.MINECRAFT_1_8.ID
}

// HasLoginPluginMessages reports whether a client on this version answers a
// login plugin request, which is how a proxy with a forwarding secret is
// asked for the login it holds. 1.13 is where the two packets appeared. A
// client before it -- 1.12.2, 1.12.1, 1.12, 1.11.1, 1.11, 1.10, 1.9.3, 1.9.2, 1.9.1, 1.9, 1.8, 1.7.6 or 1.7.2 -- has neither, and a request sent to
// one is a packet it cannot read; so a login on it goes the way one does when the
// client says it has never heard of the channel, and a proxy in front of it
// can forward only in the handshake.
func (v ProtocolVersion) HasLoginPluginMessages() bool {
	return v.ID >= ProtocolVersions.MINECRAFT_1_13.ID
}

// MaySignEncryptionChallenge reports whether a client on this version may
// answer the encryption request's challenge with a signature instead of
// encrypting it. 1.19 gave the client a profile key, signed by Mojang, and a
// client holding one -- which is every client logged into an account --
// signs the challenge under it and never encrypts it; a client without one
// encrypts it as every version does. 1.19.3 is where the key left the login,
// so from it on the challenge is always encrypted, and 1.18.2, 1.18, 1.17.1,
// 1.17, 1.16.4, 1.16.3, 1.16.2, 1.16.1, 1.16, 1.15.2, 1.15.1, 1.15, 1.14.4, 1.14.3, 1.14.2, 1.14.1, 1.14, 1.13.2, 1.13.1, 1.13, 1.12.2, 1.12.1, 1.12, 1.11.1, 1.11, 1.10, 1.9.3, 1.9.2, 1.9.1, 1.9, 1.8, 1.7.6 and 1.7.2, from before the key, have nothing to sign with and encrypt it as well:
// 1.19 and 1.19.1 are the two versions that may sign. A signature is a thing this server
// cannot check, since it does not keep the key the client sent with its
// hello, and a version that may sign is a version whose response is let
// through without the challenge: see CompleteEncryption.
func (v ProtocolVersion) MaySignEncryptionChallenge() bool {
	return IsSupportedProtocolVersion(v) && v.ID >= ProtocolVersions.MINECRAFT_1_19.ID && v.ID < ProtocolVersions.MINECRAFT_1_19_3.ID
}
