# VRChat auto rejoin tool (VRChatでホームに戻されたときに自動で戻るやつ）
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbootjp%2Fvrc_auto_rejoin_tool.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbootjp%2Fvrc_auto_rejoin_tool?ref=badge_shield)

## what is this?（これはなに？）

Due to internet issues VRChat may force users back to their Home World unintentionally. The use of this application can redirect user back to their previous World, which includes Invite+ and Invite only Sessions. This application is designed for VRChat Sleepers, which they could auto reconnect to the session they were sleeping in and not waking up alone in Home World.

The follow will trigger the Application to activate
   Change of World or session
   When VRChat.exe ended (only when [`enable_process_check: yes`] is active)

English translation [@s_mitsune](https://twitter.com/s_mitsune).


VR睡眠して朝起きるとホームに一人でいることを防ぎたくて作りました．

本ツールを起動したままで以下のいずれかが起きた場合元のインスタンスに自動で戻ります．
- インスタンスの移動
- VRChat.exeの終了
  - `enable_process_check: yes` のとき

## how to use （使い方）
事前にVRChatを立ち上げておく必要があります．


エラー時やクラッシュ時に、戻ってきたいインスタンスにいる状態で本ソフトウェアを立ち上げます．

立ち上げ後にエラーでホームワールドに飛ばされたり、クラッシュした場合本ソフトウェアが、先ほどまでいたインスタンスに入るようにVRChatを立ち上げます。


自動で戻りたいインスタンスにいる状態で本ソフトウェアを立ち上げます．  
立ち上げ後にインスタンスの移動を検出した場合は、先程までいたインスタンスに戻ろうとVRChatのlauncherを先程のインスタンスIDで立ち上げ直します．



## License
- [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbootjp%2Fvrc_auto_rejoin_tool.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbootjp%2Fvrc_auto_rejoin_tool?ref=badge_large)
- 同梱しているwavファイルは CeVIO の さとうささら を利用しています．

