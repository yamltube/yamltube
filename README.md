# YamlTube: Manage your YouTube Playlists in Yaml

Fork this repo to manage your YouTube playlists in YAML on GitHub.

### Who is this for?

To be honest, I have no idea. It seemed like a funny concept and yamltube.com was $14.17.

**Coming Soon Maybe**: `yamltube:spotify:Playlist`

### Long Term Requirements For This To Be Super Cool

- You should be able to set this up on your phone. Which means it needs to be 100% setup-able and managable from the browser. You should be able to fork
  yamltube's repo, click a link in the readme, and authenticate somehow with youtube and pulumi
- Two way sync of a playlist back into YAML. Basically you would edit playlist in the YouTube app and a PR automatically opens up on the repo with the new links

This is how you define a playlist:

```yaml
name: yamltube
runtime: yaml
description: Manage Your YouTube playlist in yaml
resources:
  makingmyway:
    type: yamltube:youtube:Playlist
    properties:
      title: Making My Way Downtown
      description: I guarantee you know these songs
      visibility: public # or private or unlisted
      videos:
        - https://www.youtube.com/watch?v=Cwkej79U3ek
        - https://www.youtube.com/watch?v=iPUmE-tne5U
        - https://www.youtube.com/watch?v=b7k0a5hYnSI
        - https://www.youtube.com/watch?v=qi7Yh16dA0w
        - https://www.youtube.com/watch?v=gte3BoXKwP0
        - https://www.youtube.com/watch?v=KU5o6M7S5nQ
        - https://www.youtube.com/watch?v=znlFu_lemsU
        # not supported yet
        # merge in another playlist into this one
        # - https://www.youtube.com/playlist?list=PLeQFt2AXw9mSQpqcBfHkufqpBsS2x4hTD
        # or this way works too, it just ignores the video
        # - https://www.youtube.com/watch?v=BdEe5SpdIuo&list=PLeQFt2AXw9mSQpqcBfHkufqpBsS2x4hTD
  rickroll:
    type: yamltube:youtube:Playlist
    properties:
      title: Rick Roll
      description: I'm sorry
      visibility: public
      videos:
        - https://www.youtube.com/watch?v=dQw4w9WgXcQ
outputs:
  # output links
  rickrollLink: https://www.youtube.com/playlist?list=${rickroll.playlistId}
  makingmywayLink: https://www.youtube.com/playlist?list=${makingmyway.playlistId}
```

## Setup Instructions

### Automated Setup

Coming Soon™️

### Manual Setup

1. Fork this repo
1. Sign up for Pulumi and install the CLI. Just `brew install pulumi` and do a `pulumi login`. Or read the [actual docs](https://www.pulumi.com/)
1. Obtain a `client_secret.json` for your account. Follow this [Guide](https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps) to get the file. (Sorry, no Pulumi program for this)
1. Save `client_secret.json` to root of this repo. (must be named this way). Do not commit this file
   1. You can verify this client_secret.json works by running `go build && ./verify`, and seeing if your playlists get printed out
   1. To get github actions support: Add the contents of the file to a GitHub Actions Secret as `YOUTUBE_CLIENT_SECRET`
1. Modify `Pulumi.yml` and add your playlists
1. Run `pulumi up`
