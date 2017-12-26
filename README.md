#chuantou 内网穿透 v1.2版本
##软件作用：可以让全世界访问家用电脑里的网站。
##原理
###client运行在家用电脑，内装有自己的网站。user为访问网站的浏览器。
![输入图片说明](http://git.oschina.net/uploads/images/2017/0313/235933_fd3a3ee6_891703.png "在这里输入图片标题")
##内网穿透v1.2版特色
 **1、多协程并发管理多个tcp链接。速度更快。** 

 **2、加入心跳包机制（20秒）。应对拔网线等极端情况** 

 **3、自定义服务器与家用电脑的监听端口** 

 **4、支持断线重连，如果没收到心跳包或者网断了会自动重连**

 **5、server端基本不需要关闭，如需要重启可以只重启客户端（client）** 

 **6、本版本为重构版本，不需要引入额外的包，只需两个文件。代码清晰易懂，并加入大量释。** 
##使用方法：
###1、配置好go语言环境，
###2、把server.go上传到公网服务器上。运行例子：go run server.go -localPort 3002 -remotePort 20012（如下图）
localPort端口为用户访问的端口，remotePort端口为与client通讯的端口。
![输入图片说明](http://git.oschina.net/uploads/images/2017/0329/103259_a77dfa0e_891703.png "在这里输入图片标题")

###3、把client.go放在家用电脑上（无公网ip，只能家用电脑80端口可以访问到本地的网站）。运行例子go run client.go -host 服务器ip -localPort 80 -remotePort 20012（如下图）
localPort端口为家用电脑网站的端口，remotePort端口为与server通讯的端口，与server端设置必须一致
![输入图片说明](http://git.oschina.net/uploads/images/2017/0329/103312_81b123e0_891703.png "在这里输入图片标题")

###4、全世界任何浏览器访问公网ip:3002即可访问到家用电脑中的网站。（如服务器ip为1.1.1.1则访问1.1.1.1:3002）（如下图）
![输入图片说明](http://git.oschina.net/uploads/images/2017/0329/103334_7bd07f0d_891703.png "在这里输入图片标题")

##补充：
1.作者初入tcp网络编程。软件还不完善，如果发现bug欢迎提交issue

 **2.另外大家觉得有帮助，可以捐助本项目请我喝杯cafe。** 
