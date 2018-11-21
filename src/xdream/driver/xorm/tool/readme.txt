./tool -db=sports_mall -debug=1 -force=true -out=/tmp/model -config=/data/home/fankxu/wk/fly_dev/src/fly/example/httpserver/conf -prefix=""
 参数说明：
 config:必填， 配置文件的根目录，目录下必须存在应用配置文件config.json
 db: 必填， 数据库名，需要在配置文件中配置
 table:可选， 需要生产model代码的表名，*表示所有表，多个表可以用逗号分隔，默认为所有表
 force:可选，是否覆盖已经存在的文件，默认不覆盖
 out:可选， 输出文件的路径，默认输出到当前目录下./model文件夹中
 pkg:可选，默认从out参数获取，取目录的最后一部分作为生成类的package
 debug:可选,  是否输出调试信息，0.不输出（默认） 1.输出调试信息
 prefix:可选，类名前缀，默认不添加前缀