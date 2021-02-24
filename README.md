# ChoyKit

Foundation Kit 

设计初衷是可以跨多个项目使用的基础工具包



### 各子包说明


包名        |  描述
------------|-----------------------------
collections | 容器和算法
datetime    | 日期相关
dotenv      | .env文件解析
fsutil      | 文件相关
mathext     | 数学扩展包  
reflectutil | 反射相关
strutil     | 字符串相关  



### 编码规范

1. API接口不能使用panic抛出错误；
2. 包命名要简短并准确，不使用下划线、驼峰，勿使用太笼统的名字（如common, base, misc），[包名规范细节](https://blog.golang.org/package-names)；
3. 本module的包之间不相互引用；
4. 不过多依赖第三方外部包；
