# YamlTube

#### YamlTube lets you manage your youtube playlists in yaml on GitHub using GitHub actions.

Just go to https://yamltube.com to create your GitHub repo. It'll create and configure a repo automatically for you.

Setting up the repo can is done entirely in your browser.

### Who is this for? Why does this exist?

No idea. I thought of the name `yamltube` and yamltube.com was available.

### What does a YouTube playlist in Yaml look like?

```yaml
youtube:
  playlists:
    - title: Never Gonna Give You Up
      description: Never Gonna Let You Down
      visibility: public # or "private", or "unlisted"
      videos:
        - https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Future Work: Spotify Playlists?

```yaml
spotify:
  playlists:
    - title: Walkin Fast
      tracks:
        - link: https://open.spotify.com/track/4w1lzcaoZ1IC2K5TwjalRP
        # or
        - title: A Thousand Miles
          artist: Vanessa Carlton
          album: Be Not Nobody
        # or
        - isrc: USIR10210955 # https://www.isrcfinder.com/
```

It would also be super cool to be able to create Apple, YouTube, Tidal(?), playlists as well.
Maybe define a single "playlist" and have it synchronized between a bunch of different accounts? Would anybody actually use that?

### I don't want to grant you credentials to my google account. Can I run this myself?

Yah. Just follow the instructions on the [yamltube/bin](https://github.com/yamltube/bin) repo.

https://yamltube.com only exists to make setup easier for those who don't want to waste their time in the GCP console.
