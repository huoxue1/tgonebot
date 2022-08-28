# tgonebot

onebot v12的telegram机器人实现


## 支持的动作

### onebot v12 标准动作
  + get_version
  + get_status
  + get_self_info
  + send_message
  + delete_message
  + get_group_info
  + get_file
  + get_group_member_info

### 扩展动作
  + set_group_ban
  + set_group_kick
  + get_commands
  + set_commands
  + edit_text_message
  + set_inline_key_board


## 支持推送的事件
+ message
  - private
  - group
  - channel
+ notice
  - group_member_increase
  - group_member_decrease

## 支持的消息段
+ text
+ image
+ video
+ audio
+ file
+ location
+ reply
+ mention