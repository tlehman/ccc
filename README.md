# ccc(1) a command line tool for catechumen

To use this command, simply type `ccc` to be greeted by the index:

```
ccc
```

## Looking up a particular paragraph

If you wanted to look up a particular paragraph, like CCC 2765, 
type `ccc 2765`:

```
$ ccc 2765
The traditional expression "the Lord's Prayer" - oratio Dominica - means that the prayer to our Father is taught and given to us by the Lord Jesus. the prayer that comes to us from Jesus is truly unique: it is "of the Lord." On the one hand, in the words of this prayer the only Son gives us the words the Father gave him:13 he is the master of our prayer. On the other, as Word incarnate, he knows in his human heart the needs of his human brothers and sisters and reveals them to us: he is the model of our prayer.
```


## Reading through the Catechism step by step

The `ccc` command can store your current position in the Catechism's text. If you want to read it as you read a book cover-to-cover, then run:

```
ccc begin
```

And the command will save your current position in ~/.cccpos

To advance that position by one, run:

```
ccc next
```
