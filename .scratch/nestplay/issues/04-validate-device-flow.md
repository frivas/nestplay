# Validate account access and playback on the target devices
<!-- Canonical local Wayfinder ticket. -->
Type: task
Label: wayfinder:task
Status: open
Assignee: none
Parent: [Decide Nestplay&apos;s proof-of-concept architecture](https://github.com/frivas/nestplay/issues/2)
Blocked by: [Establish personalized YouTube access and session lifecycle](https://github.com/frivas/nestplay/issues/3), [Select original-language metadata and Spanish subtitle sources](https://github.com/frivas/nestplay/issues/5), [Choose modular playback and realistic Mac packaging](https://github.com/frivas/nestplay/issues/4), [Accept the PoC YouTube connection experience](https://github.com/frivas/nestplay/issues/8), [Prepare metadata and subtitle provider access](https://github.com/frivas/nestplay/issues/10)

## Question

Obtain the target Mac and Apple TV model/OS, a user-mediated YouTube session, and selected representative movie/video fixtures. Run only a separately agreed bounded proof once research specifies it. Record feed persistence, seeking/subtitle playback, conversion, disk threshold behavior and resource measurements. This task unblocks architecture confirmation; it is not a production build.

## Comments

User confirmed A1625 and tvOS 26.6. Apple identifies A1625 as Apple TV HD. Target its supported 1080p output; validate higher-resolution/HDR sources only through conversion. Host observed: M1 Pro, 16 GB, macOS 26.6.2. No device playback has yet been tested.
