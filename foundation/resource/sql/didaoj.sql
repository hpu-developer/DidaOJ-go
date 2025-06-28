-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- 主机： 59.110.20.167:3306
-- 生成日期： 2025-06-27 16:52:58
-- 服务器版本： 5.7.44
-- PHP 版本： 8.2.28

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `didaoj`
--

-- --------------------------------------------------------

--
-- 表的结构 `collection`
--

CREATE TABLE `collection` (
  `id` int(11) NOT NULL,
  `title` varchar(30) DEFAULT NULL,
  `description` text,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `inserter` int(11) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int(11) DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `password` varchar(30) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `collection_member`
--

CREATE TABLE `collection_member` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题集参与人员';

-- --------------------------------------------------------

--
-- 表的结构 `collection_problem`
--

CREATE TABLE `collection_problem` (
  `id` int(11) NOT NULL,
  `problem_id` int(11) DEFAULT NULL,
  `index` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `contest`
--

CREATE TABLE `contest` (
  `id` int(11) NOT NULL,
  `title` varchar(30) NOT NULL,
  `description` text,
  `notification` varchar(100) DEFAULT NULL,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `inserter` int(11) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int(11) DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `password` varchar(30) DEFAULT NULL,
  `submit_anytime` tinyint(1) DEFAULT NULL,
  `type` tinyint(4) DEFAULT NULL,
  `score_type` tinyint(4) DEFAULT NULL,
  `lock_rank_duration` bigint(20) DEFAULT NULL,
  `always_lock` tinyint(1) DEFAULT NULL,
  `discuss_type` tinyint(4) DEFAULT NULL COMMENT '讨论类型，0正常讨论，1仅查看自己的讨论'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `contest_language`
--

CREATE TABLE `contest_language` (
  `id` int(11) NOT NULL,
  `language` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member`
--

CREATE TABLE `contest_member` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛访问权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_auth`
--

CREATE TABLE `contest_member_auth` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛管理权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_ignore`
--

CREATE TABLE `contest_member_ignore` (
  `user_id` int(11) NOT NULL,
  `id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛忽略排行成员';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_volunteer`
--

CREATE TABLE `contest_member_volunteer` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛志愿者权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_problem`
--

CREATE TABLE `contest_problem` (
  `id` int(11) NOT NULL,
  `problem_id` int(11) DEFAULT NULL,
  `index` int(11) DEFAULT NULL,
  `view_id` int(11) DEFAULT NULL,
  `score` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `discuss`
--

CREATE TABLE `discuss` (
  `id` int(11) NOT NULL,
  `title` varchar(30) DEFAULT NULL,
  `content` text,
  `view_count` int(11) DEFAULT NULL,
  `banned` tinyint(1) DEFAULT NULL,
  `problem_id` int(11) DEFAULT NULL,
  `contest_id` int(11) DEFAULT NULL,
  `judge_id` int(11) DEFAULT NULL,
  `inserter` int(11) NOT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int(11) NOT NULL,
  `modify_time` datetime DEFAULT NULL,
  `updater` int(11) NOT NULL,
  `update_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `discuss_comment`
--

CREATE TABLE `discuss_comment` (
  `id` int(11) NOT NULL,
  `discuss_id` int(11) NOT NULL,
  `content` text,
  `banned` tinyint(1) DEFAULT NULL,
  `inserter` int(11) NOT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int(11) NOT NULL,
  `modify_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `discuss_tag`
--

CREATE TABLE `discuss_tag` (
  `id` int(11) NOT NULL,
  `tag_id` int(11) NOT NULL,
  `index` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `judger`
--

CREATE TABLE `judger` (
  `key` varchar(10) NOT NULL,
  `name` varchar(20) NOT NULL,
  `max_job` int(11) DEFAULT NULL,
  `cpu_usage` float DEFAULT NULL,
  `mem_usage` bigint(20) UNSIGNED DEFAULT NULL,
  `mem_total` bigint(20) UNSIGNED DEFAULT NULL,
  `avg_message` varchar(100) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `judge_job`
--

CREATE TABLE `judge_job` (
  `id` int(11) NOT NULL,
  `problem_id` int(11) DEFAULT NULL,
  `contest_id` int(11) DEFAULT NULL,
  `contest_problem_index` int(11) DEFAULT NULL,
  `language` tinyint(1) DEFAULT NULL,
  `code` text,
  `code_length` int(11) DEFAULT NULL,
  `status` tinyint(1) DEFAULT NULL,
  `judger` varchar(10) DEFAULT NULL,
  `judge_time` datetime DEFAULT NULL,
  `task_current` int(11) DEFAULT NULL,
  `task_total` int(11) DEFAULT NULL,
  `score` int(11) DEFAULT NULL,
  `time` int(11) DEFAULT NULL,
  `memory` int(11) DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `remote_judge_id` varchar(20) DEFAULT NULL,
  `remote_account_id` varchar(20) DEFAULT NULL,
  `inserter` int(11) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `judge_job_compile`
--

CREATE TABLE `judge_job_compile` (
  `id` int(11) DEFAULT NULL,
  `message` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `problem`
--

CREATE TABLE `problem` (
  `id` int(11) NOT NULL COMMENT '问题ID',
  `key` varchar(18) NOT NULL COMMENT '问题对外标识',
  `sort` tinyint(1) DEFAULT NULL COMMENT '排序等级',
  `origin_oj` varchar(10) DEFAULT NULL COMMENT '原始OJ标识',
  `origin_id` varchar(8) DEFAULT NULL COMMENT '原始ID标识',
  `origin_url` varchar(100) DEFAULT NULL COMMENT '原始链接',
  `origin_author` varchar(255) DEFAULT NULL,
  `title` varchar(50) NOT NULL COMMENT '标题',
  `description` text NOT NULL COMMENT '描述',
  `source` varchar(120) DEFAULT NULL,
  `time_limit` int(11) DEFAULT NULL COMMENT '时间限制',
  `memory_limit` int(11) DEFAULT NULL COMMENT '内存限制',
  `judge_type` tinyint(1) DEFAULT NULL,
  `judge_md5` char(32) DEFAULT NULL,
  `inserter` int(11) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL COMMENT '创建时间',
  `modifier` int(11) DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL COMMENT '创建时间',
  `accept` int(11) DEFAULT NULL COMMENT '通过人数',
  `attempt` int(11) DEFAULT NULL COMMENT '尝试人数',
  `private` tinyint(1) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- 转存表中的数据 `problem`
--

INSERT INTO `problem` (`id`, `key`, `sort`, `origin_oj`, `origin_id`, `origin_url`, `origin_author`, `title`, `description`, `source`, `time_limit`, `memory_limit`, `judge_type`, `judge_md5`, `inserter`, `insert_time`, `modifier`, `modify_time`, `accept`, `attempt`, `private`) VALUES
(1, '1', 1, NULL, NULL, NULL, NULL, 'A + B Problem', '## 题目描述\n\nCalculate A + B .\n\n计算 A + B 。\n\n## 输入\n\nEach line will contain two integers A and B ( $-10^5<A,B<10^5$ ).\n\nProcess to end of file.\n\n每一行包含两个整数 A 和 B ( $-10^5<A,B<10^5$ )。\n\n处理到文件结束。\n\n（这意味着你的程序不能只计算一次 $A+B$ 就退出）\n\n## 输出\n\nFor each case,\n\noutput $A+B$ in one line.\n\n对于每组测试，\n\n在一行输出 $A+B$ 。\n\n（所以请注意行末换行）\n\n## 样例输入\n\n```\n1 1\n2 3\n```\n\n## 样例输出\n\n```\n2\n5\n```\n\n## 提示\n\n以下是各个语言本题的AC代码，可以用于参考。\n\n### C:\n\n```C\n#include <stdio.h>\nint main(){\n    int a,b;\n    while(scanf(\"%d %d\",&a, &b) != EOF)\n        printf(\"%d\\n\",a+b);\n	return 0;\n}\n```\n\n### C++:\n\n```cpp\n#include <iostream>\nusing namespace std;\nint main(){\n    int a,b;\n    while(cin >> a >> b)\n        cout << a+b << endl;\n    return 0;\n}\n```\n\n### Java:\n\n```Java\nimport java.util.*;\npublic class Main{\n    public static void main(String args[]){\n        Scanner cin = new Scanner(System.in);\n        int a, b;\n        while (cin.hasNext()){\n            a = cin.nextInt(); b = cin.nextInt();\n            System.out.println(a + b);\n        }\n    }\n}\n```\n\n### Python:\n\n```Python\nimport sys\nfor line in sys.stdin:\n    a = line.split()\n    print(int(a[0]) + int(a[1]))\n```\n\n### Pascal:\n\n```pascal\nprogram p1001(Input,Output); \nvar \n    a,b:Integer; \nbegin \n    while not eof(Input) do \n        begin \n            Readln(a,b); \n            Writeln(a+b); \n        end; \nend.\n```\n\n### Golang:\n\n```go\npackage main\n\nimport (\n	\"fmt\"\n)\n\nfunc main() {\n	var a, b int\n	for {\n		_, err := fmt.Scan(&a, &b)\n		if err != nil {\n			break // 输入失败（例如 EOF），退出循环\n		}\n		fmt.Println(a + b)\n	}\n}\n```\n\n### Lua\n\n```lua\nwhile true do\n    local line = io.read()\n    if not line then break end\n\n    local a, b = line:match(\"^(%-?%d+)%s+(%-?%d+)$\")\n    if a and b then\n        print(tonumber(a) + tonumber(b))\n    end\nend\n```\n\n### TypeScript\n\n```typescript\nimport * as readline from \'readline\';\n\n// 创建读取接口\nconst rl = readline.createInterface({\n  input: process.stdin,\n  output: process.stdout,\n  terminal: false\n});\n\n// 监听每一行输入\nrl.on(\'line\', (line: string) => {\n  const [a, b] = line.trim().split(/\\s+/).map(Number);\n  if (!isNaN(a) && !isNaN(b)) {\n    console.log(a + b);\n  }\n});\n\n```\n\n当然，纯JavaScript的语法也是兼容的。\n\n```javascript\nconst readline = require(\'readline\');\n\nconst rl = readline.createInterface({\n  input: process.stdin,\n  output: process.stdout,\n  terminal: false\n});\n\nrl.on(\'line\', (line) => {\n  const [a, b] = line.trim().split(/\\s+/).map(Number);\n  if (!isNaN(a) && !isNaN(b)) {\n    console.log(a + b);\n  }\n});\n```\n', 'OJ入门', 1000, 131072, 0, '4cc0f94200ecdd38b99658bfef303c06', 3, '2017-11-15 10:09:46', 3, '2025-06-19 15:39:59', 762, 1569, 0),
(2, '10', 2, NULL, NULL, NULL, NULL, 'QAQ的区间价值', '## 问题描述\n\nQAQ有一个整数序列，元素个数有$N$个，分别为$ 1，2，3 ... N-1，N $。\n\n假定用数组$a[]$来依次记录$N$个元素，那么区间$[i, j]$的价值$ = sum(i, j) * Check(i, j)$。\n\n其中$ sum(i, j) = \\sum_{k=i}^j a[k] $，表示区间$[i, j]$元素之和。\n\n$ Check(i, j) = (j - i + 1) $ & $ 1 ? 1 : -1$，表示区间长度为奇数返回$ 1 $，反之返回$ -1 $。\n\n现在，QAQ想知道所有子区间的价值之和。\n\n对于区间$ [2, 4] $而言，子区间有$ 6 $个$[2, 2], [2, 3], [2, 4], [3, 3], [3, 4], [4, 4]$。\n\n## 输入\n\n第一行输入一个整数$T$，代表有$T$组测试数据。\n\n每组数据占两行，第一行输入一个整数$N$，代表序列的元素个数。\n\n注：$ 1 \\leq T \\leq 10000，1 \\leq N \\leq 100000$。\n\n## 输出\n\n对每组测试数据，输出序列所有子区间价值之和。\n\n## 样例输入\n\n```\n3\n1\n2\n99\n```\n\n## 样例输出\n\n```\n1\n0\n2500\n```\n\n', 'CZY', 1000, 131072, 0, 'd5c92fcbd299556b159ca6bc8ab9b852', 3, '2016-10-26 10:19:28', 3, '2016-10-26 10:19:28', 84, 147, 0),
(3, '100', 3, NULL, NULL, NULL, NULL, '三国杀', '## 问题描述\n\n《三国杀》是一款热门的桌上游戏，该游戏融合了西方类似游戏的特点，并结合中国三国时期背景，以身份为线索，以卡牌为形式，合纵连横，经过一轮一轮的谋略和动作获得最终的胜利。三国杀集合历史、文学、美术等元素于一身，在中国广受欢迎。\n\n![三国杀](https://r2-oj.didapipa.com/problem/100/1750325631_97e7af95-c063-4a6a-a615-aed02ed52a81 \'三国杀\')\n\n为了感受《三国杀》的魅力，KACA分析了游戏的流程。\n\n游戏中的基础攻击卡牌是【杀】(出牌阶段，对攻击范围内除自己以外的一名角色使用，效果是对该角色造成1点伤害。)。\n\n现在KACA假设了一种情况，现在有若干名玩家进行游戏，武将皆为白板，有若干条操作，KACA想知道操作后会产生什么结果。\n\n已知攻击范围默认为1，相邻座位的玩家默认距离为1。\n\n此外还有两种马具：\n\n- +h可使当其他角色出【杀】计算与该角色距离时，始终+h。\n- -h可使当该角色出【杀】计算与其他角色距离时，始终-h。\n\n当然，一个角色最多只能装备一匹-h马和一匹+h马，若已有马具，再次装备同类型则会把之前的马具替换掉。\n\n## 输入\n\n多组测试数据。\n\n第一行是两个数字n，m(2≤n≤1000,1≤m≤1000)分别代表玩家数量和操作次数。\n\n下面m行有两种形式\n\n1. 某个玩家装备了一匹马，i P h(1≤i≤n，P为+或-，h(1≤h≤n)代表该马可提供的距离)\n2. 询问玩家A能否杀玩家B，K A B(K为一个字符，1≤A,B≤n)\n\n## 输出\n\n对于每组数据中的每次询问，若玩家A可以对玩家B出杀，则输出“No.# Can”，否则输出“No.# Cann\'t”，#代表A的编号。\n\n## 样例输入\n\n```\n5 3\nK 1 3\n1 - 1\nK 1 3\n\n```\n\n## 样例输出\n\n```\nNo.1 Cann\'t\nNo.1 Can\n\n```', 'BoilTask', 1000, 131072, 0, '622d75ea523f047a62485610d9183b6c', 3, '2016-12-25 16:44:03', 3, '2025-06-19 17:34:28', 15, 89, 0),
(4, '1000', 4, NULL, NULL, NULL, NULL, '二分搜索', '<p><span style=\"font-size: medium;\">在有序序列中查找某一元素x。</span></p>\n\n## 输入\n<p><span style=\"font-size: medium;\">首先输入一个正整数n(n&lt;=100000)，表示该序列有n个整数，然后按从小到大的顺序输入n个整数；</span></p>\n<p><span style=\"font-size: medium;\">接着是一个正整数m，表示有m次查找；</span></p>\n<p><span style=\"font-size: medium;\">最后是m个整数，表示m个要查找的整数x。</span></p>\n\n## 输出\n<p><span style=\"font-size: medium;\">对于每一个次查找，有一行输出。若序列中存在要查找的元素x，则输出元素x在序列中的序号（序号从0开始）；若序列中不存在要查找的元素x，则输出&quot;Not found!&quot;。</span></p>\n\n## 样例输入\n```\n5\n1 3 5 7 9 \n11\n-1\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n\n```\n\n\n## 样例输出\n```\nNot found!\n0\nNot found!\n1\nNot found!\n2\nNot found!\n3\nNot found!\n4\nNot found!\n\n```\n', '', 1000, 131072, 0, '0706e34f43185fd8c7355d39cd97f100', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:50:49', 0, 0, 0),
(5, '1001', 4, NULL, NULL, NULL, NULL, '简易版最长序列', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">给你一组数(未排序)，请你设计一个程序：求出里面个数最多的数。并输出这个数的长度。 <br />\n例如：给你的数是：1、 2、 3、 3、 4、 4、 5、 5、 5 、6, 其中只有6组数：1, 2, 3-3, 4-4, 5-5-5 and 6. <br />\n最长的是5那组，长度为3。所以输出3。 </font><br />\n<br />\n</span></p>\n\n## 输入\n<p><font size=\"3\" face=\"Times New Roman\">第一行为整数t((1 &le; t &le; 10))，表示有n组测试数据。 <br />\n每组测试数据包括两行，第一行位数组的长度n (1 &le; n &le; 10000)。第二行为n个整数，所有整数Mi的范围都是(1 &le; Mi&nbsp;&lt; 2^32) <br />\n<br />\n<br />\n</font></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">对应每组数据，输出个数最多的数的长度。 </font><br />\n<br />\n</span></p>\n\n## 样例输入\n```\n1\n10\n1 2 3 3 4 4 5 5 5 6\n```\n\n\n## 样例输出\n```\n3\n```\n', '', 1000, 131072, 0, 'cc827e97747857eb7c1b8fdf020b69d1', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:50:53', 0, 0, 0),
(6, '1002', 4, NULL, NULL, NULL, NULL, '字符串比较 多实例', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">比较字符串大小，但比较的规则不同字典序规则。字符比较新规则如下：</font></span><span style=\"font-size: medium\"><font face=\"Times New Roman\">A &lt; a &lt; B &lt; b &lt; &hellip;&hellip;&hellip;&hellip; &lt; Z &lt; z。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入数据包含多个测试实例，每个测试实例占两行，每一行有一个字符串（只包含大小写字母， 长度小于10000）。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">如果第一个字符串小于第二个，输出YES，否则，输出NO。 <br />\n注意：A &lt; a &lt; B &lt; b &lt; &hellip;&hellip;&hellip;&hellip; &lt; Z &lt; z。</font></span></p>\n\n## 样例输入\n```\nabc\nBbc\nAb\na\nABcef\nABce\n\n\n```\n\n\n## 样例输出\n```\nYES\nYES\nNO\n\n\n```\n', '', 1000, 131072, 0, 'b76f480b86d31be026a5a3afec8e3cfb', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:50:56', 0, 0, 0),
(7, '1003', 4, NULL, NULL, NULL, NULL, '单数变复数', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个名词英语单词，按照英语语法规则把单数变成复数。规则如下： <br />\n（1） 以辅音字母y结尾，变y为i，再加es； <br />\n（2） 以s, x, ch, sh结尾，则加es； <br />\n（3） 以元音o结尾，则加es； <br />\n（4） 其他情况加上s。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个字符串，仅含小写字母，长度不超过20。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出其对应的复数形式。</font></span></p>\n\n## 样例输入\n```\nbutterfly\n\n```\n\n\n## 样例输出\n```\nbutterflies\n\n\n```\n', '', 1000, 131072, 0, '06012948c29967e645912a7e1b38a7cd', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:03', 0, 0, 0),
(8, '1004', 4, NULL, NULL, NULL, NULL, '连续的n个1', '<div><span style=\"font-size: small\">计算机数据都是由0和1组成的，看着长长的0101001110101111011，要找出连续n个1的子串有多少个，确实麻烦，请你编程实现吧。</span></div>\n\n## 输入\n<p><span style=\"font-size: small\"><span style=\"font-family: 宋体; mso-ascii-font-family: Calibri; mso-hansi-font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA\">输入第一行为一个字符串，由</span><span lang=\"EN-US\" style=\"font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA; mso-fareast-font-family: 宋体\">0</span><span style=\"font-family: 宋体; mso-ascii-font-family: Calibri; mso-hansi-font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA\">和</span><span lang=\"EN-US\" style=\"font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA; mso-fareast-font-family: 宋体\">1</span><span style=\"font-family: 宋体; mso-ascii-font-family: Calibri; mso-hansi-font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA\">组成，长度小于</span><span lang=\"EN-US\" style=\"font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA; mso-fareast-font-family: 宋体\">1000</span><span style=\"font-family: 宋体; mso-ascii-font-family: Calibri; mso-hansi-font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA\">；输入第二行为一个正整数</span><span lang=\"EN-US\" style=\"font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA; mso-fareast-font-family: 宋体\">n</span><span style=\"font-family: 宋体; mso-ascii-font-family: Calibri; mso-hansi-font-family: Calibri; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-font-size: 12.0pt; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-bidi-language: AR-SA\">。</span></span></p>\n\n## 输出\n<div><span style=\"font-size: small\">输出一个整数，表示连续n个的1的子串的个数。</span></div>\n\n## 样例输入\n```\n0101001110101111011\n2\n```\n\n\n## 样例输出\n```\n6\n```\n', '', 1000, 131072, 0, '1f7d8b233790403672ba62f04130795b', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:10', 0, 0, 0),
(9, '1005', 4, NULL, NULL, NULL, NULL, '又是排序（指针专题）', '<p><font face=\"Times New Roman\" size=\"3\">将输入的四个整数按由大到小的顺序输出。 <br />\n已定义如下swap函数，可实现形参pa和pb所指内存单元的内容交换。请务必使用本函数实现两个变量内容的互换。 <br />\nvoid swap( int *pa, int *pb) <br />\n{ <br />\nint t; <br />\nt=*pa; *pa=*pb; *pb=t; <br />\n} <br />\n</font></p>\n\n## 输入\n<p><font face=\"Times New Roman\" size=\"3\">输入4个整数，用空格隔开。</font></p>\n\n## 输出\n<p><font face=\"Times New Roman\" size=\"3\">输出排序后的4个整数，由空格隔开。输出占一行。</font></p>\n\n## 样例输入\n```\n4 3 5 2\n```\n\n\n## 样例输出\n```\n5 4 3 2\n```\n', '', 1000, 131072, 0, 'f17d9dc2b49e716905019483fef8593b', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:16', 0, 0, 0),
(10, '1006', 4, NULL, NULL, NULL, NULL, '最大的两个数（指针专题）', '<p></p>\n<p></p>\n<p></p>\n<p></p>\n<p></p>\n<p></p>\n<p><span style=\"font-size: medium\">求n个整数中的最大值和最小值。要求定义一个函数LargestTow()，求数组a的最的两个值，分别存入形参指针pfirst和psecond所指存储单元，函数原型如下： </span></p>\n<p><span style=\"font-size: medium\">void LargestTow(int a[],int n,int *pfirst,int *psecond) <br />\n{ <br />\n/*数组a有n个元素，将数组中的最大值存入形参指针pfirst所指内存单元，将数组中第二大的值存入形参指针psecond所指内存单元。 */<br />\n}</span></p>\n<div></div>\n<div v:shape=\"_x0000_s1026\">\n<div style=\"mso-line-spacing: \'100 50 0\'; mso-margin-left-alt: 216; mso-char-wrap: 1; mso-kinsoku-overflow: 1\"></div>\n</div>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入有两行，输入第一行是一个整数n，1&lt;n&lt;=1000；第二行是n个整数，由空格隔开。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入两个整数，表示数组中最大的两个值。输出占一行。</font></span></p>\n\n## 样例输入\n```\n5\n6 3 4 9 8\n```\n\n\n## 样例输出\n```\n9 8\n```\n', '', 1000, 131072, 0, 'b704ab5e0a83574e6ffa2dfb02f3fba9', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:22', 0, 0, 0),
(11, '1007', 4, NULL, NULL, NULL, NULL, '矩阵的最大值（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">找出一个2&times;3的整数矩阵中的最大值及其行下标和列下标，要求调用函数FindMax(int p[][3], int m, int n, int *pRow, int *pCol)实现，行下标和列下标在形参中以指针的形式返回。 <br />\nvoid FindMax(int p[][3], int m, int n, int *pRow, int *pCol){ <br />\n//在m*n矩阵p中查找最大值，将其行下标存入pRow所指内存单元，将其列下标存入pCol所指内存单元 <br />\n} <br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入2行3列整数，共6个。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出3个整数，矩阵中的最大值及其行下标和列下标，数据之间用空格隔开。测试数据保证最大值唯一。</font><br />\n<br />\n</span></p>\n\n## 样例输入\n```\n100 3 6\n0 87 65\n```\n\n\n## 样例输出\n```\n100 0 0\n```\n', '', 1000, 131072, 0, 'a20b365dd657bf69d6605ba1afe07a45', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:28', 0, 0, 0),
(12, '1008', 4, NULL, NULL, NULL, NULL, '字符串长度（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">编写一函数len，求一个字符串的长度，注意该长度不计空格。要求用字符指针实现。在主函数中输入字符串，调用该len函数后输出其长度。 <br />\nint len(char *sp) <br />\n{ <br />\n//实现sp所指串的长度，不计空格。 <br />\n} </font><br />\n<br />\n</span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个字符串，以回车结束，长度不超过100。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出一个整数，单独占一行。</font></span></p>\n\n## 样例输入\n```\nWhat day is today?\n\n```\n\n\n## 样例输出\n```\n15\n```\n', '', 1000, 131072, 0, '1bb261e6d36e60c3b6401ac36aedec57', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:33', 0, 0, 0),
(13, '1009', 4, NULL, NULL, NULL, NULL, '循环移动（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有n个整数，使前面各数顺序向后移动k个位置，移出的数再从开头移入。输出移动后的数组元素。 <br />\n题目没有告诉你n的范围，希望你读入n之后用malloc()函数动态申请内存空间，不要提前定义数组的大小。不要偷懒哦。 <br />\n另外要求定义并使用函数ringShift() <br />\nvoid ringShift(int *a, int n, int k) <br />\n{ <br />\n//循环移动后的数值仍然存入数组a中 <br />\n}</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入分3行，第一行是一个整数n，第二行是n个整数，用空格隔开。第三行是一个整数k。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出n个整数，由空格隔开。输出占一行。</font></span></p>\n\n## 样例输入\n```\n6\n1 2 3 4 5 6\n2\n```\n\n\n## 样例输出\n```\n5 6 1 2 3 4\n```\n', '', 1000, 131072, 0, '52fe5ab8fb051c36b1db66b465bf24fb', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:41', 0, 0, 0),
(14, '101', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[1.3.1]三数的最大值', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">求3个数中的最大值。</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试为三个整数a,b,c(-1000&lt;a,b,c&lt;1000)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">输出a,b,c中的最大值。</span>\r\n</p>\n\n## 样例输入\n\n```\n1 2 3\r\n\n```\n\n## 样例输出\n\n```\n3\r\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, '0b2da13aae89e7cf1f78eafd005c759a', 3, '2016-11-09 22:09:43', 3, '2016-11-09 22:09:43', 395, 950, 0),
(15, '1010', 4, NULL, NULL, NULL, NULL, '亲和串（字符串）', '<p><span style=\"font-size: small\">判断亲和串。亲和串的定义是这样的：给定两个字符串s1和s2,如果能通过s1循环移位，使s2包含在s1中，那么我们就说s2 是s1的亲和串。<br />\n</span></p>\n\n## 输入\n<div class=\"panel_content\"><span style=\"font-size: small\">本题有多组测试数据，每组数据的第一行包含输入字符串s1,第二行包含输入字符串s2，s1与s2的长度均小于100000。<br />\n</span></div>\n\n## 输出\n<div class=\"panel_content\"><span style=\"font-size: small\">如</span><span style=\"font-size: small\">果s2是s1的亲和串，则输出&quot;yes&quot;，反之，输出&quot;no&quot;。每组测试的输出占一行。<br />\n</span></div>\n\n## 样例输入\n```\nAABCD\nCDAA\nASD\nASDF\nab\naba\n\n```\n\n\n## 样例输出\n```\nyes\nno\nno\n\n```\n', '', 1000, 131072, 0, '34edb9b76c12b53f643bc4c0c7cc3b32', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:44', 0, 0, 0),
(16, '1011', 4, NULL, NULL, NULL, NULL, '在线判题（字符串）', '<p><span style=\"font-family: \'Courier New\';\"><span style=\"font-size: medium;\">LittleTom开发了一个在线判题系统，判题系统需要把用户提交上来的代码编译成可执行文件，然后运行。而用户会提交什么样的代码是无法预知的，所以LittleTom做了充分的准备，比如阻止解题程序访问文件系统、阻止解题程序访问注册表、阻止解题程序修改系统设置、阻止解题程序关闭系统、阻止解题程序超限或非法使用内存、阻止解题程序的运行时间超过设定时间等。这些工作LitteTom都已完成。<br />\n&nbsp;&nbsp; 还有一个待解决的问题是判断解题程序的正确性。判题系统需要把解题程序产生的输出文件和正确的输出文件进行比较，如果两个文件完全相同，则判题系统返回&ldquo;Accepted&rdquo;，否则，如果两个文件除了空白符(空格\' \', 制表符\'\\t\',&nbsp;或 回车符\'\\n\')之外其余内容都相同，则判题系统返回&ldquo;Presentation Error&rdquo;，否则判题系统返回&ldquo;Wrong Answer&rdquo;。<br />\n给定两个文件，一个代表正确输出，一个代表用户的解题程序的输出，你的任务是计算判题系统应该返回什么信息。</span></span></p>\n\n## 输入\n<p><span style=\"font-family: \'Courier New\';\"><span style=\"font-size: medium;\">输入包含多组测试实例。第一行输入一个整数T表示测试实例的个数。然后是T组输入。每组输入有两部分：一个代表正确输出，一个代表用户的解题程序的输出。都以&ldquo;START&rdquo;开始，以&ldquo;END&rdquo;结束，在&ldquo;START&rdquo;和&ldquo;END&rdquo;之间的是需要判断的数据部分。</span></span></p>\n\n## 输出\n<p><span style=\"font-family: \'Courier New\';\"><span style=\"font-size: medium;\">对于每一个测试实例，有一行输出，输出判题系统应该返回的结果：Accepted、Presentation Error或Wrong Answer。</span></span></p>\n\n## 样例输入\n```\n4\nSTART\n1 + 2 = 3\nEND\nSTART\n1+2=3\nEND\nSTART\n1 + 2 = 3\nEND\nSTART\n1 + 2 = 3\nEND\n\nSTART\n1 + 2 = 3\nEND\nSTART\n1 + 2 = 4\nEND\nSTART\n1 + 2 = 3\nEND\nSTART\n1	+	2	=	3\nEND\n\n```\n\n\n## 样例输出\n```\nPresentation Error\nAccepted\nWrong Answer\nPresentation Error\n\n```\n', '', 1000, 131072, 0, 'caff24cd60130b7e23f1b03a0fcaba1a', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:49', 0, 0, 0),
(17, '1012', 4, NULL, NULL, NULL, NULL, '实数的小数部分（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">读入一个实数，输出该实数的小数部分，小数部分若多余的末尾0，请去掉。如输入111111.12345678912345678900 <br />\n则输出0.123456789123456789。若去掉末尾0之后小数部分为0，则输出&ldquo;No decimal part&rdquo;。注意该实数的位数不超过100位。 <br />\n请定义并使用如下函数。 <br />\nchar *decimal(char *p) <br />\n{ <br />\n将字符串p表示的实数的自小数点开始的小数部分存入一个字符串，并由函数返回，若p为&ldquo;123.456&rdquo;，则返回的字符串为&ldquo;.456&rdquo;。若小数部分为0,返回空指针NULL。 <br />\n} <br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个实数。不超过100位。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出小数部分，输出占一行。</font></span></p>\n\n## 样例输入\n```\n111111.12345678900012345678900\n```\n\n\n## 样例输出\n```\n0.123456789000123456789\n```\n', '', 1000, 131072, 0, '89c6fe1ef3ff36bebdd917a4d0eaebb0', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:51:54', 0, 0, 0),
(18, '1013', 4, NULL, NULL, NULL, NULL, '实数取整（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">读入一个实数，输出实数的整数部分。注意该实数的位数不超过100位。输入的整数部分可能含有不必要的前导0，输出时应去掉，当然，若整数部分为0，则该0不能去掉。如输入0023.56732，输出应为23，而不是0023；0.123对应的输出应为0。当然输入也可能不含小数部分。 <br />\n要求定义并使用rounding()函数，原型如下： <br />\nchar *rounding(char *p) <br />\n{ <br />\n//将字符串p表示的实数取整后生成新的字符串，并由函数返回 <br />\n} <br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个实数.</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出整数部分。</font></span></p>\n\n## 样例输入\n```\n0012345678900.56789\n\n```\n\n\n## 样例输出\n```\n12345678900\n\n```\n', '', 1000, 131072, 0, '55fafbb3bc4d50088e7e1c94d4b93d74', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:00', 0, 0, 0),
(19, '1014', 4, NULL, NULL, NULL, NULL, '逆转数（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">任意给你一个整数，这个数可能很大（最长不超过100位），你能求出它的逆转数吗？ <br />\n逆转数定义如下： <br />\n1.一个末尾没有0的整数，它的逆转数就是各位数字逆序输出； <br />\n2.一个负数的逆转数仍是负数； <br />\n3.一个末尾有0的整数，它的逆转数如同下例： <br />\nreverse (1200) = 2100 <br />\nreverse (-56) = -65 <br />\n要求定义并使用如下函数： <br />\nvoid reverse(char *str) <br />\n{ <br />\n//函数求出str的逆转数并存入str。 <br />\n} <br />\n<br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一个长整数str，不超过100位，输入的整数不含前导0。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出str的逆转数。输出占一行。</font></span></p>\n\n## 样例输入\n```\n-123456789000\n\n```\n\n\n## 样例输出\n```\n-987654321000\n\n```\n', '', 1000, 131072, 0, 'bc78295f61054f7fc07c74a52ec03205', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:06', 0, 0, 0),
(20, '1015', 4, NULL, NULL, NULL, NULL, '账单（指针专题）', '<p><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">每到月末，小明就会对这个月的支出账单进行整理和统计。如今电脑已经普及大学校园，所以小明想让电脑帮忙做这件事情。聪明的你就为小明编一个程序来完成这件事情吧。</span></p>\n\n## 输入\n<p><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">多实例测试。首先输入一个整数ncase，表示测试实例的个数。每个测试实例的输入如下：</span></p>\n<p><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">第一行是整数</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\">n (n&lt;100)</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">。然后是</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\">n</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">行的账单信息，每一行由事物的名字</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\">name</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">和对应的花费</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\">c</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">组成，长度不超过</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\">200</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">。中间会有一个或多个空格，而每一行的开头和结尾没有空格。</span><span lang=\"EN-US\" style=\"line-height: 130%; font-family: Verdana; font-size: 10.5pt; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN; mso-fareast-font-family: 宋体\"> 0.0 &lt; c &lt; 1000.0</span><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">。</span></p>\n\n## 输出\n<p><span style=\"line-height: 130%; font-family: 宋体; font-size: 10.5pt; mso-ascii-font-family: Verdana; mso-hansi-font-family: Verdana; mso-bidi-font-size: 10.0pt; mso-bidi-font-family: \'Times New Roman\'; mso-font-kerning: 1.0pt; mso-bidi-language: AR-SA; mso-ansi-language: EN-US; mso-fareast-language: ZH-CN\">每个测试实例对应一行输出，输出总的花费，小数点后保留一位数字。</span></p>\n\n## 样例输入\n```\n2\n1\nBuy books 62.28\n3\nApple 2.3\nBuy clothes for   girl friend 260.5\nGo to  cinema 30\n\n```\n\n\n## 样例输出\n```\n62.3\n292.8\n\n```\n', '', 1000, 131072, 0, 'aa0754bd5358ae7b11ee2a806667486f', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:09', 0, 0, 0),
(21, '1016', 4, NULL, NULL, NULL, NULL, '大整数（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入3个大整数，位数不超过100位，按从小到大的顺序输出这三个整数。要求定义并使用如下函数比较两个大整数的大小。 <br />\nint cmp(char *a,char *b) <br />\n{ <br />\n//若大整数a大于b，返回1； <br />\n//若a小于b，返回-1； <br />\n// 若a与b相等，返回0 <br />\n}</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入有3行，每行输入一个大整数，位数不超过100位，输入不含前导0。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出3行，即排序后的3个大整数。</font></span></p>\n\n## 样例输入\n```\n1234567890123456789\n99999999999999\n111111111111111\n\n```\n\n\n## 样例输出\n```\n99999999999999\n111111111111111\n1234567890123456789\n```\n', '', 1000, 131072, 0, 'dbaa92a07a4b71d75a888bf453b56d5d', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:15', 0, 0, 0),
(22, '1017', 4, NULL, NULL, NULL, NULL, '最长字符串（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入多个字符串，输出最长字符串。要求定义并使用函数maxLenStr()， <br />\nvoid maxLenStr(char *str[], int n, int *max) <br />\n{ <br />\n从字符串数组str中找出最长的一个字符串，并将其下标存入形参指针max所指内存。 <br />\n}</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入有多行，每行一个字符串，每个字符串长度不超过80，输入最多不超过100行，用****作为结束输入的标志，该行输入不用处理。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出最长的一个字符串。</font></span></p>\n\n## 样例输入\n```\nL love C programming\nACM/ICPC\nstudy hard\n****\n\n```\n\n\n## 样例输出\n```\nL love C programming\n\n```\n', '', 1000, 131072, 0, 'f0a4de3eab37b559690905222ccdd4f0', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:20', 0, 0, 0),
(23, '1018', 4, NULL, NULL, NULL, NULL, '加密（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">将一段明文加密。加密的规则如下：将每个字符的ascii码的值减去24作为每个字符加密后的值，例如\'a\'的ascii码的值为97，那么加密后就变成了73。&quot;73&quot;就是\'a\'的密文。现在请你编写程序，对一段文字加密。请定义并使用如下函数： <br />\nvoid encrypt(char *plain, char *cipher) <br />\n{ <br />\n//把原文字符串plain加密后存入字符串cipher <br />\n} </font><br />\n</span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一串字符串，只包含数字和字母，最长为200.</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出加密后的字符串。</font></span></p>\n\n## 样例输入\n```\nzero12\n\n```\n\n\n## 样例输出\n```\n987790872526\n\n```\n\n\n## 提示\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">直接输出密文当然是最简单的方法，不过本题要求你将密文存入字符串(以备程序的其它模块使用）。</font></span></p>', '', 1000, 131072, 0, 'cbc960cda20b1dad1297a8e1946971f8', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:25', 0, 0, 0),
(24, '1019', 4, NULL, NULL, NULL, NULL, '矩阵边界和（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">给定一个m行n列的二维矩阵，求其四周边元素和。1&lt;=m、n&lt;=100000，可能是1行100000列，也可能是10000行50列，但保证矩阵元素不多于500000。你可能不能预定义数组的大小了，你要学会使用动态内存分配哦。你可以动态申请m*n个内存单元，然后用一维数组来存储二维数组，二维数组元素a[i][j]对应一维数组a[i*n+j]，i、j均从0开始。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入第一行是m和n，然后是一个m行n列的矩阵。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出一个整数，表示矩阵所有边界元素的和。</font></span></p>\n\n## 样例输入\n```\n3 4\n1 2 3 4 \n5 6 7 8 \n9 5 4 6 \n```\n\n\n## 样例输出\n```\n47\n```\n', '', 1000, 131072, 0, '4074c7d585e8d1177e549847e9490c4b', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:39', 0, 0, 0),
(25, '102', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[1.3.2]求1~n相加', '## 问题描述\n\n请输出函数 $   f \\left( n \\right) = \\sum _{1}^{n} x $ 的结果。\n\n也就是1+2+3+……+n的值\n\n## 输入\n\n每次测试有一个正整数n(n<1000)。\n\n## 输出\n\n输出所求结果。\n\n## 样例输入\n\n```\n100\n```\n\n## 样例输出\n\n```\n5050\n```\n\n', '贾宗璞许合利主编', 1000, 131072, 0, '5cf33396c6f49768e64f4ef09774897b', 3, '2016-11-09 22:10:20', 3, '2025-06-07 10:58:51', 352, 485, 0),
(26, '1020', 4, NULL, NULL, NULL, NULL, '密码解密（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有加密当然也有解密啦。那我们来解密吧。已知明文中只有数字和字母，且加密的规则如下：将每个字符的ascii码的值减去24作为每个字符加密过后的密文，例如\'a\'的ascii码的值为97，那么加密后就变成了73。&quot;73&quot;就是\'a\'的密文。所以，若密文是&ldquo;757392&rdquo;，则解密后的原文是&ldquo;cat&rdquo;。现在请你编写程序，将一段密文解密为原文。<br />\n请定义并使用如下函数 <br />\nvoid decrypt(char *cipher, char *plain) <br />\n{ <br />\n//将密文cipher解密后将明文存入plain <br />\n} </font><br />\n</span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入一串字符串，最长为200。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出解密后的原文。</font></span></p>\n\n## 样例输入\n```\n987790872526\n\n```\n\n\n## 样例输出\n```\nzero12\n```\n', '', 1000, 131072, 0, '7cbc58681d05d5924a719084fe2e3363', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:43', 0, 0, 0),
(27, '1021', 4, NULL, NULL, NULL, NULL, '长整数排序（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">长整数排序。输入n 然后输入n个位数不超过100位的大整数，输入的整数可能含有前导0。将这n个长整数排序后输出，输出不含前导0。 <br />\nint greater(char *s1, char *s2) <br />\n{ <br />\n若s1指向的整数大于s2指向的整数，返回1; <br />\n若s1指向的整数小于s2指向的整数，返回-1; <br />\n若s1指向的整数等于s2指向的整数，返回0; <br />\n} <br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入第一行是一个正整数n（n&lt;=10)，接下来n行是n个大整数,均非负。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出n行，为n个升序排序后的大整数。</font></span></p>\n\n## 样例输入\n```\n3\n012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789\n54213456565\n113456789456231545674632132156465132123156456423132156465461321654878976464654132132156455\n\n```\n\n\n## 样例输出\n```\n54213456565\n12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789\n113456789456231545674632132156465132123156456423132156465461321654878976464654132132156455\n\n```\n', '', 1000, 131072, 0, '47fe9762d586522f48e666ea685e45bb', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:48', 0, 0, 0),
(28, '1022', 4, NULL, NULL, NULL, NULL, '矩阵转置（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">给定一个m行n列的二维矩阵，输出其转置矩阵。1&lt;=m、n&lt;=100000，可能是1行100000列，也可能是10000行50列。你可能不能预定义数组的大小了，你要学会使用动态内存分配哦。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入第一行是m和n，然后是一个m行n列的矩阵。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出转置后的矩阵</font></span></p>\n\n## 样例输入\n```\n3 4\n1 2 3 4 \n5 6 7 8 \n9 5 4 6 \n\n\n```\n\n\n## 样例输出\n```\n1 5 9\n2 6 5\n3 7 4\n4 8 6\n```\n', '', 1000, 131072, 0, '0a715fe0c319874e204e544935102e01', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:52:56', 0, 0, 0),
(29, '1023', 4, NULL, NULL, NULL, NULL, '查找最大字符串（指针专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">从键盘上输入多个字符串（每个串不超过5个字符且没有空格），用&rdquo;*****&rdquo;作为串输入结束的标记。从所输入的若干字符串中，找出一个最大的串，并输出该串。要求最大串的查找通过调用编写的函数实现 <br />\nvoid find(char *name[], int n, int *p) <br />\n{ <br />\n//在包含n个字符串的二维字符数组name中，查找值最大的字符串，将其下标存入指针p所指内存单元 <br />\n} </font><br />\n<br />\n</span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">一行输入一个字符串，输入多行</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出一个字符串，单独占一行。 </font><br />\n</span></p>\n\n## 样例输入\n```\nzzzdf\nfdsg\nadff\nrtrt\n*****\n\n```\n\n\n## 样例输出\n```\nzzzdf\n```\n', '', 1000, 131072, 0, '1b81dcd2595a264c38d01ef96ac50fc8', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:00', 0, 0, 0),
(30, '1024', 4, NULL, NULL, NULL, NULL, '按要求排序（指针专题）', '<p><span style=\"font-size: small\">输入n和n个整数，然后按要求排序，若输入1，请输出升序排序序列；若输入2，请输出降序排序序列，若输入3，请输出按绝对值升序排序序列。要求程序结构如下，请完善程序。</span></p>\n<p>void sort(int a[], int n, int (*cmp)());<br />\nint CmpAsc(int x, int y); /*按升序要求判断两元素是否逆序*/<br />\nint CmpDec(int x, int y); /*按降序要求判断两元素是否逆序*/<br />\nint CmpAbsAsc(int x, int y);&nbsp; /*按绝对值升序要求判断两元素是否逆序*/<br />\nint main(void)<br />\n{<br />\n&nbsp;&nbsp; int a[10],i,n;&nbsp;<br />\n&nbsp;&nbsp; int slt;<br />\n&nbsp; <br />\n&nbsp;/*读入n和n个整数，存入数组a*/<br />\n&nbsp;&nbsp;<br />\n&nbsp;&nbsp; /*读入用户的选择，存入slt; */</p>\n<p><br />\n&nbsp;&nbsp;&nbsp; switch(slt)<br />\n&nbsp;&nbsp; {<br />\n&nbsp;&nbsp;&nbsp;&nbsp; case 1:&nbsp;&nbsp; sort(a, n, CmpAsc); break;<br />\n&nbsp;&nbsp;&nbsp;&nbsp; case 2:&nbsp;&nbsp; sort(a, n, CmpDec); break;<br />\n&nbsp;&nbsp;&nbsp;&nbsp; case 3:&nbsp;&nbsp; sort(a, n, CmpAbsAsc);break;<br />\n&nbsp;&nbsp; }</p>\n<p>&nbsp;&nbsp; /*输出排序后的数组*/</p>\n<p><br />\n&nbsp;&nbsp; }</p>\n<p>void sort(int a[], int n, int (*cmp)())<br />\n{ <br />\n&nbsp;/*对数组a排序，排序原则由cmp指定，若cmp为真，表示两元素逆序*/</p>\n<p>&nbsp;}</p>\n<p>int CmpAsc(int x, int y)<br />\n{<br />\n&nbsp;//如果x&gt;y返回1，否则返回0；<br />\n&nbsp;<br />\n}</p>\n<p>int CmpDec(int x, int y)<br />\n{<br />\n&nbsp;//如果x&lt;y返回1，否则返回0；<br />\n&nbsp;}<br />\n&nbsp;<br />\nint CmpAbsAsc(int x, int y)<br />\n{</p>\n<p>//如果abs(x)&gt;abs(y)返回1，否则返回0<br />\n}<br />\n</p>\n<p></p>\n\n## 输入\n<p>输入第一行是一个正整数n;</p>\n<p>输入第二行是n个整数；</p>\n<p>输入第三行是一个1~3的整数slt，表示用户的排序要求。</p>\n\n## 输出\n<p>输出n个整数。若用户的排序选择是1，则输出升序排序后的n个整数；若用户的排序选择是2，则输出降序排序后的n个整数；若用户的排序选择是3，则输出按绝对值升序排序后的n个整数；输出占一行，数据之间用空格隔开。</p>\n\n## 样例输入\n```\n5\n2 -3 1 5 4\n2\n\n```\n\n\n## 样例输出\n```\n5 4 2 1 -3\n\n```\n', '', 1000, 131072, 0, '11a5b07be0013cb3a35f8e3925f3673b', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:07', 0, 0, 0),
(31, '1025', 4, NULL, NULL, NULL, NULL, '单词数', '<p><span style=\"widows: 2; text-transform: none; text-indent: 0px; display: inline !important; font: medium \'Times New Roman\'; white-space: normal; orphans: 2; float: none; letter-spacing: normal; color: rgb(0,0,0); word-spacing: 0px; -webkit-text-size-adjust: auto; -webkit-text-stroke-width: 0px\">统计一篇文章里不同单词的总数。</span></p>\n\n## 输入\n<p><span style=\"widows: 2; text-transform: none; text-indent: 0px; display: inline !important; font: medium \'Times New Roman\'; white-space: normal; orphans: 2; float: none; letter-spacing: normal; color: rgb(0,0,0); word-spacing: 0px; -webkit-text-size-adjust: auto; -webkit-text-stroke-width: 0px\">有多组数据，每组一行，每组就是一篇小文章。每篇小文章都是由小写字母和空格组成，没有标点符号，遇到#时表示输入结束。</span></p>\n\n## 输出\n<p><span style=\"widows: 2; text-transform: none; text-indent: 0px; display: inline !important; font: medium \'Times New Roman\'; white-space: normal; orphans: 2; float: none; letter-spacing: normal; color: rgb(0,0,0); word-spacing: 0px; -webkit-text-size-adjust: auto; -webkit-text-stroke-width: 0px\">每组只输出一个整数，其单独成行，该整数代表一篇文章里不同单词的总数。</span></p>\n\n## 样例输入\n```\nyou are my friend    \n#\n\n```\n\n\n## 样例输出\n```\n4\n\n```\n', '', 1000, 131072, 0, '6cbe68ec00e0de2b296c29b90dbcc388', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:09', 0, 0, 0),
(32, '1026', 4, NULL, NULL, NULL, NULL, '带参宏定义(函数专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">从键盘输入三个字符，用空格隔开，使用带参宏定义1中SWAP，将三个字符按从大到小的顺序排序输出。 <br />\n宏定义1：#define SWAP(a, b, t) { t=a; a=b; b=t; } <br />\n<br />\n请尝试，如果用宏定义2中的SWAP，主函数需要如何修改才能得到正确结果？ <br />\n宏定义2：#define SWAP(a, b, t) t=a; a=b; b=t; <br />\n</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入三个字符，用空格隔开</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出占一行，包含三个字符，用空格隔开</font></span></p>\n\n## 样例输入\n```\nw a q\n```\n\n\n## 样例输出\n```\nw q a\n```\n', '', 1000, 131072, 0, '31403101199e9a2334cee65f7029c3ca', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:16', 0, 0, 0),
(33, '1027', 4, NULL, NULL, NULL, NULL, '成绩统计（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">从键盘输入若干个学生的信息，每个学生信息包括学号、姓名、3门课的成绩，计算每个学生的总分，输出总分最高的学生的信息。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数，然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及三个整数，表示语文、数学、英语三门课成绩，数据之间用空格隔开。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出总成绩最高的学生的学号、姓名、及三门课成绩，用空格隔开。若有多个最高分，只输出第一个。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 89 78 95\n541207010189 Wangli 85 87 99\n541207010190 Fangfang 85 68 76\n\n```\n\n\n## 样例输出\n```\n541207010189 Wangli 85 87 99\n\n\n```\n', '', 1000, 131072, 0, 'de6d52ac9de7ef0d8f57084e1de16cea', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:21', 0, 0, 0),
(34, '1028', 4, NULL, NULL, NULL, NULL, '谁的年龄最小（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">设计一个结构体类型，包含姓名、出生日期。其中出生日期又包含年、月、日三部分信息。输入n个好友的信息，输出年龄最小的好友的姓名和出生日期。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=10)，表示好友人数，然后输入n行，每行包含一个好友的信息：姓名（不超过8位）以及三个整数，分别表示出生日期的年月日。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出年龄最小的好友的姓名和出生日期，用空格隔开，出生日期的输出格式见输出样例。</font></span></p>\n\n## 样例输入\n```\n3\nZhangling 1983 2 4\nWangliang 1983 12 11\nFangfang 1985  6 1\n```\n\n\n## 样例输出\n```\nFangfang 1985-06-01\n```\n', '', 1000, 131072, 0, '03ef11f705762f7125e77b500a73da41', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:27', 0, 0, 0),
(35, '1029', 4, NULL, NULL, NULL, NULL, '按出生日期排序（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">送人玫瑰手有余香，小明希望自己能带给他人快乐，于是小明在每个好友生日的时候发去一份生日祝福。小明希望将自己的通讯录按好友的生日排序排序，这样就查看起来方便多了，也避免错过好友的生日。为了小明的美好愿望，你帮帮他吧。小明的好友信息包含姓名、出生日期。其中出生日期又包含年、月、日三部分信息。输入n个好友的信息，按生日的月份和日期升序输出所有好友信息。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=10)，表示好友人数，然后输入n行，每行包含一个好友的信息：姓名（不超过8位），以及三个整数，分别表示出生日期的年月日。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">按过生日的先后（月份和日期）输出所有好友的姓名和出生日期，用空格隔开，出生日期的输出格式见输出样例。 </font><br />\n</span></p>\n\n## 样例输入\n```\n3\nZhangling 1985 2 4\nWangliang 1985 12 11\nFangfang 1983  6 1\n```\n\n\n## 样例输出\n```\nZhangling 1985-02-04\nFangfang 1983-06-01\nWangliang 1985-12-11\n\n\n```\n', '', 1000, 131072, 0, 'f1b1abd3fea85bcd13f4edba64cc4da0', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:32', 0, 0, 0),
(36, '103', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[1.3.3]一元二次方程的根', '## 问题描述\n\n对于方程 $ a{x}^{2} + b x + c = 0 $，请判断方程有无实根。\n\n若有两个不等的实根，输出“Two”。\n\n若有两个相等的实根，输出“One”。\n\n若有两个复根，输出“None”。\n\n## 输入\n\n每次测试有三个实数($ -1000 < a,b,c < 1000 , a \\neq 0 $)。\n\n## 输出\n\n根据要求输出结果。\n\n## 样例输入\n\n```\n1 0 -1\n\n```\n\n## 样例输出\n\n```\nTwo\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, 'a0970f749f9ccc9ed448c94a126ba15c', 3, '2016-11-09 22:10:47', 3, '2025-06-23 13:44:47', 272, 1106, 0),
(37, '1030', 4, NULL, NULL, NULL, NULL, '平面点排序（一）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">平面上有n个点，坐标均为整数。请按与坐标原点（0,0）距离的远近将所有点排序输出。可以自己写排序函数，也可以用qsort库函数排序。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入有两行，第一行是整数n（1&lt;=n&lt;=10)，接下来有n行，每行一对整数（每对整数对应一个点）。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出排序后的所有点，格式为(u,v)，每个点后有一个空格。测试数据保证每个点到原点的距离都不同。</font></span></p>\n\n## 样例输入\n```\n4\n1 3 \n2 5 \n1 4 \n4 2\n\n```\n\n\n## 样例输出\n```\n(1,3) (1,4) (4,2) (2,5) \n\n```\n', '', 1000, 131072, 0, 'aedde2cac6e75d808b1cd9f8a135ba95', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:38', 0, 0, 0),
(38, '1031', 4, NULL, NULL, NULL, NULL, '平面点排序（二）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">平面上有n个点，坐标均为整数。横坐标相同时按纵坐标排序，否则按横坐标排序。本题要求用结构体存储坐标，再进行排序。先升序排序输出，再降序排序输出，可以自己写排序函数，也可以用qsort库函数排序。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入有两行，第一行是整数n(1&lt;=n&lt;=100)，表示下一行有n对整数（每对整数对应一个点）输入，每个数据后有一个空格。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出有两行，即排序后的点，格式为(u,v)，每个点后有一个空格。第一行升序排序结果，第二行降序排序结果。</font></span></p>\n\n## 样例输入\n```\n4\n1 3\n2 5\n1 4\n4 1\n\n\n```\n\n\n## 样例输出\n```\n(1,3) (1,4) (2,5) (4,1) \n(4,1) (2,5) (1,4) (1,3) \n\n```\n', '', 1000, 131072, 0, 'e68a003a32f0694ce162264169ece6a6', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:44', 0, 0, 0),
(39, '1032', 4, NULL, NULL, NULL, NULL, '数星星（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">一天，小明坐在院子里数星星，Gardon就出了个难题给她：Gardon在天空画出了一个矩形区域，让他输入矩形区域里有多少颗星星，仁慈的上帝还为他标出了每个星星的坐标。但小明数着数着就看花了眼，但他的原则是：宁愿多数一次，不可错过一个。如果小明把他数过的星星的坐标都告诉你，你能否帮他进行排重处理（如果两个星星的坐标位置相同，则视为一个星星），计算出星星的个数。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(n&lt;=300)，接下来的n对整数，每对表示小明数过的一个星星的位置（星星的坐标在-10000到10000之间）。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出星星的个数。</font></span></p>\n\n## 样例输入\n```\n5\n0 3\n1 5\n1 1\n0 1\n1 5\n\n```\n\n\n## 样例输出\n```\n4\n```\n', '', 1000, 131072, 0, 'b92036db7d4d51c4c5bf923fddcd18ee', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:50', 0, 0, 0),
(40, '1033', 4, NULL, NULL, NULL, NULL, '奖学金', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">某校发放奖学金共5种，获取条件各不同： <br />\n1.AK奖学金，每人8000，期末平均成绩&gt;80，且在本学期发表论文大于等于1篇； <br />\n2.AC奖学金，每人4000，期末平均成绩&gt;85，且班级评议成绩&gt;80； <br />\n3.成绩优秀奖，每人2000，期末平均成绩&gt;90； <br />\n4.西部奖学金，每人1000，期末平均成绩&gt;85的西部省份学生； <br />\n5.班级贡献奖，每人850，班级评议成绩&gt;80的学生干部。 <br />\n只要符合条件就可以得奖，一人可兼得多项奖学金。例：某生，期末平均成绩87，班级评议成绩82，且是学生干部，则可同时获得AC奖学金和班级贡献奖，奖金总数4850。现给出若干学生的姓名、期末平均成绩、班级评议成绩、是否学生干部、是否西部省份学生、发表论文数。计算哪个同学获得的奖金总数最高？有多个最高值则输出第一个出现的。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">第一行是一个整数n（1 &lt;= n &lt;= 10），表示学生的总数。接下来的n行,每行是一位学生的数据，从左向右依次是姓名，期末平均成绩，班级评议成绩，是否是学生干部，是否是西部省份学生，以及发表的论文数。姓名是由大小写英文字母组成的长度不超过20的字符串（不含空格）；期末平均成绩和班级评议成绩都是0到100之间的整数（包括0和100）；是否是学生干部和是否是西部省份学生分别用一个字符表示，Y表示是，N表示不是；发表的论文数是0到10的整数（包括0和10）。每两个相邻数据项之间用一个空格分隔。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出包括三行，第一行是获得最多奖金的学生的姓名，第二行是这名学生获得的奖金总数。如果有两位或两位以上的学生获得的奖金最多，输出他们之中在输入文件中出现最早的学生的姓名。第三行是这n个学生获得的奖学金的总数。</font></span></p>\n\n## 样例输入\n```\n4\nYaoLin 87 82 Y N 0\nChenRuiyi 88 78 N Y 1\nLiXin 92 88 N N 0\nZhangQin 83 87 Y N 1\n\n```\n\n\n## 样例输出\n```\nChenRuiyi\n9000\n28700\n\n```\n', '', 1000, 131072, 0, 'd9b258f9507f7391b881336b8616029f', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:55', 0, 0, 0),
(41, '1034', 4, NULL, NULL, NULL, NULL, '棒棒糖（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">新年快到了，计算机学院新年晚会正在筹备中，今年有一个新创意：来参加晚会的所有学生都有礼物（一根棒棒糖）。老师把买棒棒糖的任务交给小明了，并指定了要买的棒棒糖的品牌和品种。俗话说得好，货比三家不吃亏。小明来到了商店，看了各个店铺里这种棒棒糖的价格，不仅如此，他还记住了每个店铺的存货量。已知小明打算购买n根棒棒糖，问他最少要花多少钱？</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">第一行输入一个整数n，表示要购买的棒棒糖数量；第二行是一个整数m(1&lt;=m&lt;=10)，表示明明考察过的店铺的数量；接下来m行，每行两个数，表示该店铺中棒棒糖的价格和数量，价格为一实数(最多两位小数），数量为一整数。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出一个实数，保留两位小数，表示最小费用。</font></span></p>\n\n## 样例输入\n```\n100\n4\n0.5 50\n0.33 30\n1 80\n0.6 40\n\n```\n\n\n## 样例输出\n```\n46.90\n```\n', '', 1000, 131072, 0, 'b3d81b96f741e7ed479df52099903c07', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:53:59', 0, 0, 0),
(42, '1035', 4, NULL, NULL, NULL, NULL, '选票统计（一）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">1483. 某单位进行选举，有5位候选人：zhang、wang、zhao、liu、miao。编写程序，统计每人所得的票数。要求每人的信息里包括两部分：name和votes，分别描述姓名和所得票数。每个人的信息用一个结构体来表示，5个人的信息使用结构体数组。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n，表示一张选票，接下来n行，每行是一个由小写英文字母组成的字符串，表示以上5个候选人之一。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出5行，按zhang、wang、zhao、liu、miao的顺序输出5个候选人的姓名和所得票数，用空格隔开。</font></span></p>\n\n## 样例输入\n```\n8\nzhang\nwang\nzhang\nzhao\nwang\nliu\nwang\nwang\n\n```\n\n\n## 样例输出\n```\nzhang 2\nwang 4\nzhao 1\nliu 1\nmiao 0\n\n```\n', '', 1000, 131072, 0, '0894bd87e0b9e2efb559e15e5be42c7d', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:03', 0, 0, 0),
(43, '1036', 4, NULL, NULL, NULL, NULL, '选票统计（二）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">计算机与通信工程学院2012届学生会主席投票选举工作定于1月6日在电教楼前隆重举行。本次投票将采用电脑统计选票的方式，当投票选举正式开始后，同学们将排队一一走到电脑前，投上自己神圣的一票：在电脑前输入一个姓名，然后输入回车表示确认。 <br />\n当所同学投票结束，工作人员只需要输入一个&rdquo;#&rdquo;并回车确认，表示投票结束，电脑立即显示出得票最高的同学姓名，该同学将当选为新一届计算机与通信工程学院学生会主席。 <br />\n选举大会的准备工作正在紧张进行，编程统计投票的工作就交给你了。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">本题为单实例测试。输入包含多行，每行是一个由小写英文字母组成的字符串，表示一个姓名，遇到#时表示输入结束。 你可以假定，参加投票学生的人数不超过500人，每个学生姓名字符串的长度小于20。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出为一个字符串&mdash;&mdash;当选为学生会主席的学生姓名 </font><br />\n</span></p>\n\n## 样例输入\n```\nwanglin\nliming\nwanglin\nwanglin\nfangfang\nliming\n#\n\n\n\n```\n\n\n## 样例输出\n```\nwanglin\n```\n', '', 1000, 131072, 0, 'f7308fab8f168969c07c9c733aef9264', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:15', 0, 0, 0),
(44, '1037', 4, NULL, NULL, NULL, NULL, '查询记录（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有一学生成绩表，包括学号、姓名、3门课程成绩。请实现如下查找功能：输入一个学生的学号，输出该学生学号、姓名、4门课程成绩 <br />\n</font></span><font size=\"3\" face=\"Times New Roman\"><br />\n</font></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数； <br />\n然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及3个整数，表示3门课成绩，数据之间用空格隔开。 <br />\n最后一行输入一个学号num </font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">若学号num存在，输出该学生的学号、姓名、3门课程成绩；若该学号不存在，则输出&ldquo;Not Found&rdquo;。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 89 78 95 \n541207010189 Wangli 85 87 99 \n541207010189 Fangfang 85 68 76 \n541207010188\n\n```\n\n\n## 样例输出\n```\n541207010188 Zhangling 89 78 95\n```\n', '', 1000, 131072, 0, '0c07c10f109d1f23eb8b61405f4722c0', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:22', 0, 0, 0),
(45, '1038', 4, NULL, NULL, NULL, NULL, '添加记录（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有一学生成绩表，包括学号、姓名、3门课程成绩。已知该成绩表按学号升序排序。请编程实现，添加一个新的学生信息，且使成绩表仍按学号有序；若待添加的学号与已有学号重复，则输出错误信息，拒绝添加。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数； <br />\n然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及3个整数，表示3门课成绩，数据之间用空格隔开。 <br />\n最后一行输入一个待添加的学生信息，包括学号、姓名和3门课成绩</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">若待添加的学号与已有学号重复，则输出只有一行&ldquo;error!&rdquo;；否则，输出n+1行，即添加后的成绩单信息。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 78 95 55\n541207010189 Wangli 87 99 88\n541207010191 Fangfang 68 76 75\n541207010190 Lilei 68 79 82\n```\n\n\n## 样例输出\n```\n541207010188 Zhangling 78 95 55\n541207010189 Wangli 87 99 88\n541207010190 Lilei 68 79 82\n541207010191 Fangfang 68 76 75\n\n\n```\n', '', 1000, 131072, 0, 'f6d848cdcd8b1d58933bac67108845c8', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:31', 0, 0, 0);
INSERT INTO `problem` (`id`, `key`, `sort`, `origin_oj`, `origin_id`, `origin_url`, `origin_author`, `title`, `description`, `source`, `time_limit`, `memory_limit`, `judge_type`, `judge_md5`, `inserter`, `insert_time`, `modifier`, `modify_time`, `accept`, `attempt`, `private`) VALUES
(46, '1039', 4, NULL, NULL, NULL, NULL, '删除记录（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有一学生成绩表，包括学号、姓名、3门课程成绩。请实现如下删除功能：输入一个学生的学号，删除该学生的所有信息。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数； <br />\n然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及3个整数，表示3门课成绩，数据之间用空格隔开。 <br />\n最后一行输入一个学号num。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">若要删除的学号不存在，则输出&ldquo;error!&rdquo;；否则，输出删除该学生后的所有记录。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 78 95 55\n541207010189 Wangli 87 99 88\n541207010190 Fangfang 68 76 75\n541207010188\n```\n\n\n## 样例输出\n```\n541207010189 Wangli 87 99 88\n541207010190 Fangfang 68 76 75\n\n```\n', '', 1000, 131072, 0, 'e36b4e34e39c6f53ca2406b99849519c', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:38', 0, 0, 0),
(47, '104', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[2.3.1]储存长度', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">求各种类型数据的储存长度。</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">为了简化程序，用一个数字来代表一种数据类型：</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">1:int</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">2:</span><span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">float</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">3:double</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">4:char</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试有一个整数(1~4)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">请仿照示例输出格式进行输出(byte不用加s)</span>\r\n</p>\n\n## 样例输入\n\n```\n1\n```\n\n## 样例输出\n\n```\nint : 4 byte\n```', '贾宗璞许合利主编', 1000, 131072, 0, 'f7008e63c6f19dd1bad211c661841840', 3, '2016-11-09 22:11:11', 3, '2016-11-09 22:11:11', 245, 576, 0),
(48, '1040', 4, NULL, NULL, NULL, NULL, '单科成绩排序（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有一学生成绩表，包括学号、姓名、3门课程成绩。请按要求排序输出：若输入1，则按第1门课成绩降序输出成绩表，若输入为i（1&lt;=i&lt;=3)，则按第i门课成绩降序输出成绩表。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数； <br />\n然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及3个整数，表示3门课成绩，数据之间用空格隔开。 <br />\n最后一行输入一个整数i，表示要求按第i门课成绩降序排序输出，若该门课成绩相同，则按学号升序。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出按第i门课降序排序的结果，格式见样例。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 89 78 95\n541207010189 Wangli 85 87 99\n541207010190 Fangfang 85 68 76\n1\n\n```\n\n\n## 样例输出\n```\n541207010188 Zhangling 89 78 95\n541207010189 Wangli 85 87 99\n541207010190 Fangfang 85 68 76\n\n```\n', '', 1000, 131072, 0, '52733e741364a5f74ee60e98c31e0cb8', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:45', 0, 0, 0),
(49, '1041', 4, NULL, NULL, NULL, NULL, '总成绩排序（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">有一学生成绩表，包括学号、姓名、3门课程成绩。请按如下规则排序：按总成绩降序排序，若总成绩相同，则按姓名升序排序。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数n(1&lt;=n&lt;=100)，表示学生人数； <br />\n然后输入n行，每行包含一个学生的信息：学号（12位）、姓名（不含空格且不超过20位），以及3个整数，表示3门课成绩，数据之间用空格隔开。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出排序后的成绩单，格式见输出样例。</font></span></p>\n\n## 样例输入\n```\n3\n541207010188 Zhangling 89 78 95\n541207010189 Wangli 85 87 99\n541207010190 Fangfang 89 88 85\n```\n\n\n## 样例输出\n```\n541207010189 Wangli 85 87 99 271\n541207010190 Fangfang 89 88 85 262\n541207010188 Zhangling 89 78 95 262\n\n```\n', '', 1000, 131072, 0, '948fa896ecdf07956df1a6ccbc64af53', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:51', 0, 0, 0),
(50, '1042', 4, NULL, NULL, NULL, NULL, '猴子选大王（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">n只猴子围坐成一个圈，按顺时针方向从1到n编号。然后从1号猴子开始沿顺时针方向从1开始报数，报到m的猴子出局，再从刚出局猴子的下一个位置重新开始报数，如此重复，直至剩下一个猴子，它就是大王。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入两个整数n和m,1&lt;=m&lt;=n&lt;=100。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输出猴王的编号 </font></span></p>\n\n## 样例输入\n```\n8 3\n```\n\n\n## 样例输出\n```\n7\n```\n', '', 1000, 131072, 0, '74c297b094d7ed61e28289bc384a5ef3', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:54:57', 0, 0, 0),
(51, '1043', 4, NULL, NULL, NULL, NULL, '数星星（二）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">一天，小明坐在院子里数星星，Gardon就出了个难题给他，让他数数天上的星星最多有多少个是在同一条直线上的。天上的星星太多了，小明马上就看花了眼，你能写个程序来帮他计算么？</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先输入一个整数N(N&lt;=300)，接下来的N对数每对表示一个星星的位置（星星的坐标在-10000到10000之间，精确到小数点后1位）。没有两个星星会在同一个位置。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">一个整数，表示一条直线上最多星星的数目。</font></span></p>\n\n## 样例输入\n```\n5\n0 0\n1 0\n1 1\n0 1\n0.5 0.5\n\n```\n\n\n## 样例输出\n```\n3\n```\n', '', 1000, 131072, 0, '4a5725ff94b225b5a3c54ffa80169ceb', 3, '2025-05-28 22:04:17', 3, '2025-05-29 17:55:03', 0, 0, 0),
(52, '1044', 4, NULL, NULL, NULL, NULL, '考试排名（一）（结构体专题）', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">今天浙大研究生复试的上机考试跟传统笔试的打分规则相似，总共有n道题，每道题有对应分值，最后计算总成绩。现给定录取分数线，请你写程序找出最后通过分数线的考生，并将他们的成绩按降序打印。</font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">第1行给出考生人数N ( 1&lt;= N&lt;=100 )、考题数M (1&lt;=M&lt;=10 )、分数线（正整数）G； <br />\n第2行排序给出第1题至第M题的正整数分值； <br />\n以下N行，每行给出一名考生的准考证号（长度不超过20的字符串）、该生解决的题目总数m、以及这m道题的题号 <br />\n（题目号由1到M）。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">首先在第1行输出不低于分数线的考生人数n，随后n行按分数从高到低输出上线考生的考号与分数，其间用1空格分隔。若有多名考生分数相同，则按他们考号的升序输出。</font></span></p>\n\n## 样例输入\n```\n4 5 25\n10 10 12 13 15\nCS004 3 5 1 3\nCS003 5 2 4 1 3 5\nCS002 2 1 2\nCS001 3 2 3 5\n\n```\n\n\n## 样例输出\n```\n3\nCS003 60\nCS001 37\nCS004 37\n\n```\n', '', 1000, 131072, 0, 'ad65e78ba58c14c35e23ff1b09832d51', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:09', 0, 0, 0),
(53, '1045', 4, NULL, NULL, NULL, NULL, '考试排名（二）(结构体专题)', '<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">ACM 国际大学生程序设计竞赛，英文全称：ACM International Collegiate Programming Contest（ACM-ICPC 或 ICPC）是由美国计算机协会（ACM）主办的，一项旨在展示大学生创新能力、团队精神和在压力下编写程序、分析和解决问题能力的年度竞赛。经过30多年的发展，ACM 国际大学生程序设计竞赛已经发展成为最具影响力的大学生计算机竞赛。一般就简称为 ACM 竞赛了（虽然照理来说应该简称成 ICPC 才对）。 <br />\nACM 竞赛的计分规则如下： <br />\n首先按照过题数排名，过题数相同的队伍按照罚时排名（罚时小的队伍排在前面），如果罚时也相同则认为名次相同（名次相同时在排名表上队伍 id 较小的队伍列在前面）。 <br />\n对于罚时的计算。队伍总体的罚时等于该队各题的罚时之和。对于某题的罚时，如果这道题最后没有通过（没有正确提交），则这题的罚时为 0，否则这道题的罚时为：从比赛开始到该题第一次正确提交经过的时间 + 第一次通过之前的错误提交次数 * 20 分钟。 <br />\n例如：某次考试一共8题（A，B，C，D，E，F，G，H），每个人做的题都在对应的题号下有个数量标记，负数表示该学生在该题上有过的错误提交次数，但到现在还没有AC，正数表示AC所耗的时间，如果正数a跟上一对括号，里面有个整数b，那就表示该学生提交该题AC了，耗去了时间a，同时，曾经错误提交了b次，因此对于下述输入数据： </font></span></p>\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">则其排名从高到低应该是这样的： <br />\nJosephus 5 376 <br />\nJohn 4 284 <br />\nAlice 4 352 <br />\nSmith 3 167 <br />\nBob 2 325 <br />\nBush 0 0 </font></span></p>\n\n## 输入\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">输入数据的第一行是考试题数n（1&le;n&le;12），每行数据描述一个学生的用户名（不多于10个字符的字串）以及对所有n道题的答题现状，其描述采用问题描述中的数量标记的格式，见上面的表格，提交次数总是小于100，AC所耗时间总是小于1000。 参加考试的人数不超过100人。</font></span></p>\n\n## 输出\n<p><span style=\"font-size: medium\"><font face=\"Times New Roman\">将这些学生的考试现状，输出一个实时排名。实时排名显然先按AC题数的多少排，多的在前，再按时间分的多少排，少的在前，如果凑巧前两者都相等，则按名字的字典序排，小的在前。每个学生占一行，输出名字（10个字符宽，左对齐），做出的题数（2个字符宽，右对齐）和时间分（4个字符宽，右对齐）。名字、题数和时间分相互之间有一个空格。 </font></span></p>\n\n## 样例输入\n```\n8\nSmith	  -1	-16	8	0	0	120	39	0\nJohn	  116	-2	11	0	0	82	55(1)	0\nJosephus  72(3)	126	10	-3	0	47	21(2)	-2\nBush	  0	-1	-8	0	0	0	0	0\nAlice	  -2	67(2)	13	-1	0	133	79(1)	-1\nBob	  0	0	57(5)	0	0	168	-7	0\n\n```\n\n\n## 样例输出\n```\nJosephus    5  376\nJohn        4  284\nAlice       4  352\nSmith       3  167\nBob         2  325\nBush        0    0\n\n```\n', '', 1000, 131072, 0, '01f04621258ff986216b5c3c25acfde3', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:16', 0, 0, 0),
(54, '1046', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题1.5', '<p>请参照本章例题，编写一个C程序，输出以下信息：</p>\n<p>**************************</br>\n&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Very&nbsp;&nbsp;&nbsp; Good!</br>\n**************************</p>\n<p>数＊号可看出，Very前面9空格，Good前面&hellip;&hellip;</p>\n<p>*也是输出的一部分，别光打印Very Good!</p>\n\n## 输入\n<p>无需输入</p>\n\n## 输出\n<p>**************************</br>\n&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Very&nbsp;&nbsp;&nbsp; Good!</br>\n**************************</p>\n\n## 样例输出\n```\n**************************\n         Very    Good!\n**************************\n```\n', '', 1000, 131072, 0, 'f75dde95db05c412c7e008bb91f48f93', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:19', 0, 0, 0),
(55, '1047', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题1.6', '编写一个程序，输入a、b、c三个值，输出其中最大值。\n\n## 输入\n一行数组，分别为a b c\n\n## 输出\na b c其中最大的数\n\n## 样例输入\n```\n10 20 30\n```\n\n\n## 样例输出\n```\n30\n```\n', '', 1000, 131072, 0, '2b3b783ec5598cb55d5cd53d1308b4c6', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:21', 0, 0, 0),
(56, '1048', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题3.7', '<p>要将&quot;China&quot;译成密码，译码规律是：用原来字母后面的第4个字母代替原来的字母．例如，字母&quot;A&quot;后面第4个字母是&quot;E&quot;．&quot;E&quot;代替&quot;A&quot;。因此，&quot;China&quot;应译为&quot;Glmre&quot;。请编一程序，用赋初值的方法使cl、c2、c3、c4、c5五个变量的值分别为，&rsquo;C&rsquo;、&rsquo;h&rsquo;、&rsquo;i&rsquo;、&rsquo;n&rsquo;、&rsquo;a&rsquo;，经过运算，使c1、c2、c3、c4、c5分别变为&rsquo;G&rsquo;、&rsquo;l&rsquo;、&rsquo;m&rsquo;、&rsquo;r&rsquo;、&rsquo;e&rsquo;，并输出。</p>\n\n## 输入\n<p>China</p>\n\n## 输出\n<p>加密后的China</p>\n\n## 样例输入\n```\nChina\n```\n\n\n## 样例输出\n```\nGlmre\n```\n\n\n## 提示\nso easy', '', 1000, 131072, 0, '88db1ba53b1ca89697c3ad283d538da0', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:22', 0, 0, 0),
(57, '1049', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题4.8', '<p>设圆半径r，圆柱高h 求圆周长C1、圆面积Sa、圆球表面积Sb、圆球体积Va、圆柱体积Vb。用scanf输入数据，输出计算结果，输出时要求文字说明，取小数点后两位数字。请编程序。 PI＝3.14</p>\n\n## 输入\n<p>两个浮点数，r和h</p>\n\n## 输出\n圆周长C1、圆面积Sa、圆球表面积Sb、圆球体积Va、圆柱体积Vb。\n保留两位小数，每个结果后换行。\n\n## 样例输入\n```\n1.5 3\n```\n\n\n## 样例输出\n```\nC1=9.42\nSa=7.07\nSb=28.26\nVa=14.13\nVb=21.20\n\n\n```\n', '', 1000, 131072, 0, 'ef749bb599f2e91aeb404519bb458df7', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:25', 0, 0, 0),
(58, '105', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[2.3.2]英里转换', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">从键盘输入一个公里数，屏幕输出其英里数。</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;line-height:27px;font-family:\'Microsoft YaHei\';\">每次测试有一个实数m( 0&nbsp;</span><span style=\"font-size:18px;line-height:27px;font-family:\'Microsoft YaHei\';\">&lt;= m &lt;= 1000&nbsp;</span><span style=\"font-size:18px;line-height:27px;font-family:\'Microsoft YaHei\';\">)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">根据要求输出结果（保留三位小数）。</span><span style=\"font-size:18px;\"></span>\r\n</p>\n\n## 样例输入\n\n```\n1.60934\r\n\n```\n\n## 样例输出\n\n```\n1.000\r\n\n```\n\n## 提示\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">1英里 = 1.60934公里</span>\r\n</p>', '贾宗璞许合利主编', 1000, 131072, 0, '4bd223b80a88eeb5c07f39c0ab7896f7', 3, '2016-11-09 22:12:34', 3, '2016-11-09 22:12:34', 267, 380, 0),
(59, '1050', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题4.9', '输入一个华氏温度，要求输出摄氏温度。公式为  \nc=5(F-32)/9  \n输出要求有文字说明，取位2小数。\n\n## 输入\n一个华氏温度，浮点数\n\n## 输出\n摄氏温度，浮点两位小数\n\n## 样例输入\n```\n-40\n```\n\n\n## 样例输出\n```\nc=-40.00\n```\n\n\n## 提示\n零下40度，可以不问是？氏', '', 1000, 131072, 0, '77a1748af02bb174ce1b51bb2500de7c', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:28', 0, 0, 0),
(60, '1051', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题5.4', '有三个整数a b c,由键盘输入，输出其中的最大的数。\n\n## 输入\n一行数组，分别为a b c\n\n## 输出\na b c其中最大的数\n\n## 样例输入\n```\n10 20 30\n```\n\n\n## 样例输出\n```\n30\n```\n\n\n## 提示\nmax ?\nif(a>b) ?\n\nint main(){\n\n  return 0;\n}', '', 1000, 131072, 0, 'c511ed6a622a8905b64345eef6be3b94', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:30', 0, 0, 0),
(61, '1052', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题5.5', '<p>有一个函数<br />\ny={ x&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; x&lt;1<br />\n&nbsp;&nbsp; &nbsp;| 2x-1&nbsp;&nbsp; 1&lt;=x&lt;10<br />\n&nbsp;&nbsp;&nbsp; \\ 3x-11&nbsp; x&gt;=10</p>\n<p>写一段程序，输入x，输出y</p>\n\n## 输入\n<p>一个数x</p>\n\n## 输出\n<p>一个数y</p>\n\n## 样例输入\n```\n14\n\n```\n\n\n## 样例输出\n```\n31\n```\n\n\n## 提示\n使用函数', '', 1000, 131072, 0, 'd503aa130ed18d8bb3dcca7a90c33da2', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:33', 0, 0, 0),
(62, '1053', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题5.6', '给出一百分制成绩，要求输出成绩等级‘A’、‘B’、‘C’、‘D’、‘E’。\n90分以上为A\n80-89分为B\n70-79分为C\n60-69分为D\n60分以下为E\n\n\n## 输入\n一个整数0－100以内\n\n## 输出\n一个字符，表示成绩等级\n\n## 样例输入\n```\n90\n```\n\n\n## 样例输出\n```\nA\n```\n\n\n## 提示\n分段函数返回字符', '', 1000, 131072, 0, 'e1db78de848c5814b5ea8d047ee241f0', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:35', 0, 0, 0),
(63, '1054', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题5.7', '给出一个不多于5位的整数，要求\n1、求出它是几位数\n2、分别输出每一位数字\n3、按逆序输出各位数字，例如原数为321,应输出123\n\n## 输入\n一个不大于5位的数字\n\n## 输出\n三行\n第一行 位数\n第二行 用空格分开的每个数字，注意最后一个数字后没有空格\n第三行 按逆序输出这个数\n\n## 样例输入\n```\n12345\n```\n\n\n## 样例输出\n```\n5\n1 2 3 4 5\n54321\n```\n\n\n## 提示\n哈姆雷特：数字还是字符？这是一个问题！', '', 1000, 131072, 0, '4b13b2a1520033754aff839877ee9d0e', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:38', 0, 0, 0),
(64, '1055', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题5.8', '<p>企业发放的奖金根据利润提成。利润低于或等于100000元的，奖金可提10%;<br />\n利润高于100000元，低于200000元（100000&lt;I&le;200000）时，低于100000元的部分按10％提成，高于100000元的部分，可提成 7.5%;<br />\n200000&lt;I&le;400000时，低于200000元部分仍按上述办法提成，（下同），高于200000元的部分按5％提成；<br />\n400000&lt;I&le;600000元时，高于400000元的部分按3％提成；600000&lt;I&le;1000000时，高于600000元的部分按1.5%提成；<br />\nI&gt;1000000时，超过1000000元的部分按1%提成。从键盘输入当月利润I,求应发奖金总数。</p>\n\n## 输入\n<p>一个整数，当月利润。</p>\n\n## 输出\n<p>一个整数，奖金。</p>\n\n## 样例输入\n```\n900\n```\n\n\n## 样例输出\n```\n90\n```\n\n\n## 提示\n用Switch要比用if的看起来更清晰。', '', 1000, 131072, 0, 'b07a822d94292642dd96e3c09a80ad24', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:40', 0, 0, 0),
(65, '1056', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.1', '输入两个正整数m和n，求其最大公约数和最小公倍数。\n\n## 输入\n两个整数\n\n## 输出\n最大公约数，最小公倍数\n\n## 样例输入\n```\n5 7\n```\n\n\n## 样例输出\n```\n1 35\n```\n', '', 1000, 131072, 0, '1f664ef5257c422e34be2df15c5b2492', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:43', 0, 0, 0),
(66, '1057', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.2', '输入一行字符，分别统计出其中英文字母、空格、数字和其他字符的个数。\n\n## 输入\n一行字符\n\n## 输出\n统计值\n\n## 样例输入\n```\naklsjflj123 sadf918u324 asdf91u32oasdf/.\';123\n\n```\n\n\n## 样例输出\n```\n23 16 2 4\n```\n', '', 1000, 131072, 0, 'fcc8ff50770c645773dc4670ad70d642', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:45', 0, 0, 0),
(67, '1058', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.3', '<p>求Sn=a+aa+aaa+&hellip;+aa&hellip;aaa（有n个a）之值，其中a是一个数字。 例如：2+22+222+2222+22222（n=5），n由键盘输入。a固定等于2；</p>\n\n## 输入\n<p>n</p>\n\n## 输出\n<p>和</p>\n\n## 样例输入\n```\n5\n```\n\n\n## 样例输出\n```\n24690\n```\n', '', 1000, 131072, 0, '896af8174d70baaedcfd2793298b274f', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:48', 0, 0, 0),
(68, '1059', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.4', '<p>求Sn=1!+2!+3!+4!+5!+&hellip;+n!之值，其中n是一个数字。</p>\n\n## 输入\n<p>n</p>\n\n## 输出\n<p>和</p>\n\n## 样例输入\n```\n5\n```\n\n\n## 样例输出\n```\n153\n```\n\n\n## 提示\n<p>&nbsp;long long</p>', '', 1000, 131072, 0, 'ec66c19cbd331025314e6b50816d2568', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:50', 0, 0, 0),
(69, '106', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[3.3.1]商和余数', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">编写程序，输入两个整数，求出它们的商数和余数并进行输出。</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试有两个整数a,b(0&lt;a,b&lt;1000)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">分别输出它们的商数和余数。</span>\r\n</p>\n\n## 样例输入\n\n```\n3 2\r\n\n```\n\n## 样例输出\n\n```\n1 1\r\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, 'c6bebb54a8b76b458411ee99d025e772', 3, '2016-11-09 22:12:56', 3, '2016-11-09 22:12:56', 288, 508, 0),
(70, '1060', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.5', '求以下三数的和,保留2位小数\n1~a之和\n1~b的平方和\n1~c的倒数和\n\n## 输入\na b c\n\n## 输出\n1+2+...+a + 1^2+2^2+...+b^2 + 1/1+1/2+...+1/c\n\n## 样例输入\n```\n100 50 10\n```\n\n\n## 样例输出\n```\n47977.93\n```\n', '', 1000, 131072, 0, 'c924c41a1ada5bd76633374205341900', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:52', 0, 0, 0),
(71, '1061', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.6', '打印出所有\"水仙花数\"，所谓\"水仙花数\"是指一个三位数，其各位数字立方和等于该本身。\n例如：153是一个水仙花数，因为153=1^3+5^3+3^3。\nOutput:<pre>\n153\n???\n???\n??? </pre>\n\n## 输入\n无\n\n## 输出\n所有的水仙花数，从小的开始。\n每行一个', '', 1000, 131072, 0, '02a64244341bf34e80c845bdd9fa9a63', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:54', 0, 0, 0),
(72, '1062', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.7', '<p>一个数如果恰好等于它的因子之和，这个数就称为&quot;完数&quot;。 例如，6的因子为1、2、3，而6=1+2+3，因此6是&quot;完数&quot;。 编程序找出N之内的所有完数，并按下面格式输出其因子：</p>\n\n## 输入\n<p>N</p>\n\n## 输出\n<p>? its factors are ? ? ?</p>\n\n## 样例输入\n```\n1000\n```\n\n\n## 样例输出\n```\n6 its factors are 1 2 3 \n28 its factors are 1 2 4 7 14 \n496 its factors are 1 2 4 8 16 31 62 124 248 \n```\n', '', 10000, 131072, 0, '648b98120bc30b80e4bdb35b4151ff81', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:56', 0, 0, 0),
(73, '1063', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.8', '有一分数序列：  \n2/1 3/2 5/3 8/5 13/8 21/13......\n求出这个数列的前N项之和，保留两位小数。  \n\n## 输入\nN\n\n## 输出\n数列前N项和\n\n## 样例输入\n```\n10\n```\n\n\n## 样例输出\n```\n16.48\n```\n', '', 1000, 131072, 0, '8f23661cb5198b665fd7ba4938e7d8b1', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:55:59', 0, 0, 0),
(74, '1064', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.9', '一球从M米高度自由下落，每次落地后返回原高度的一半，再落下。\n它在第N次落地时反弹多高？共经过多少米？\n保留两位小数\n\n## 输入\nM N\n\n## 输出\n它在第N次落地时反弹多高？共经过多少米？\n保留两位小数，空格隔开，放在一行\n\n## 样例输入\n```\n1000 5\n```\n\n\n## 样例输出\n```\n31.25 2875.00\n```\n', '', 1000, 131072, 0, '00299a5c7489b49800fa85d22e31747b', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:01', 0, 0, 0),
(75, '1065', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.10', '猴子吃桃问题。猴子第一天摘下若干个桃子，当即吃了一半，还不过瘾，又多吃了一个。\n第二天早上又将剩下的桃子吃掉一半，又多吃一个。以后每天早上都吃了前一天剩下的一半零一个。\n到第N天早上想再吃时，见只剩下一个桃子了。求第一天共摘多少桃子。\n\n## 输入\nN\n\n## 输出\n桃子总数\n\n## 样例输入\n```\n10\n```\n\n\n## 样例输出\n```\n1534\n```\n', '', 1000, 131072, 0, 'd47465bd61b0ffcb01918361ebb27eff', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:04', 0, 0, 0),
(76, '1066', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题6.11', '用迭代法求 。求平方根的迭代公式为：  \nX[n+1]=1/2(X[n]+a/X[n])\n要求前后两次求出的得差的绝对值少于0.00001。\n输出保留3位小数\n\n## 输入\nX\n\n## 输出\nX的平方根\n\n## 样例输入\n```\n4\n```\n\n\n## 样例输出\n```\n2.000\n```\n', '', 1000, 131072, 0, '3c722d54123e213e70001641b2cb5e87', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:07', 0, 0, 0),
(77, '1067', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题7.1', '用筛法求之N内的素数。 \n\n## 输入\nN\n\n## 输出\n0～N的素数\n\n## 样例输入\n```\n100\n```\n\n\n## 样例输出\n```\n2\n3\n5\n7\n11\n13\n17\n19\n23\n29\n31\n37\n41\n43\n47\n53\n59\n61\n67\n71\n73\n79\n83\n89\n97\n```\n\n\n## 提示\n数组大小动态定义？函数？', '', 1000, 131072, 0, '7bc609b415cc2e41bc040ca4bfb506a1', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:09', 0, 0, 0),
(78, '1068', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题7.2', '用选择法对10个整数从小到大排序。  \n\n## 输入\n10个整数\n\n## 输出\n排序好的10个整数\n\n## 样例输入\n```\n4 85  3 234 45 345 345 122 30 12\n```\n\n\n## 样例输出\n```\n3\n4\n12\n30\n45\n85\n122\n234\n345\n345\n```\n', '', 1000, 131072, 0, '67749e70eaa0882ffa6cd88f0777bbfb', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:12', 0, 0, 0),
(79, '1069', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题7.3', '求一个3×3矩阵对角线元素之和。  \n\n## 输入\n矩阵\n\n## 输出\n主对角线 副对角线 元素和\n\n## 样例输入\n```\n1 2 3\n1 1 1\n3 2 1\n```\n\n\n## 样例输出\n```\n3 7\n```\n', '', 1000, 131072, 0, 'ada407f9dd3c5c42c86821209a72fc9c', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:15', 0, 0, 0),
(80, '107', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[3.3.2]三数平均值', '## 问题描述\n\n编写程序，读入三个双精度数，求它们的平均值并保留此平均值小数点后一位数，对小数点后第二位进行四舍五入，最后输出结果。\n\n## 输入\n\n每次测试有三个双精度数a,b,c(-1000<a,b,c<1000)。\n\n## 输出\n\n求a,b,c的平均值并保留一位小数。\n\n## 样例输入\n\n```\n1 2 3\n\n```\n\n## 样例输出\n\n```\n2.0\n\n```\n', '贾宗璞许合利主编', 1000, 131072, 0, 'a2f8b1206e887585e7327235e71e08b2', 3, '2016-11-09 22:16:31', 3, '2016-11-09 22:16:31', 287, 527, 0),
(81, '1070', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题7.4', '已有一个已排好的9个元素的数组，今输入一个数要求按原来排序的规律将它插入数组中。\n\n## 输入\n第一行，原始数列。\n第二行，需要插入的数字。\n\n## 输出\n排序后的数列\n\n## 样例输入\n```\n1 7 8 17 23 24 59 62 101\n50\n```\n\n\n## 样例输出\n```\n1\n7\n8\n17\n23\n24\n50\n59\n62\n101\n\n```\n', '', 1000, 131072, 0, '63afe8ab0d4c19f1c2fbbf0d66f78138', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:17', 0, 0, 0),
(82, '1071', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.1', '写两个函数，分别求两个整数的最大公约数和最小公倍数，用主函数调用这两个函数，并输出结果两个整数由键盘输入。\n\n## 输入\n两个数\n\n## 输出\n最大公约数 最小公倍数\n\n## 样例输入\n```\n6 15\n```\n\n\n## 样例输出\n```\n3 30\n```\n', '', 1000, 131072, 0, '2de2d9ff7e25791dac61413eb3b74b99', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:19', 0, 0, 0),
(83, '1072', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.2', '求方程 的根，用三个函数分别求当b^2-4ac大于0、等于0、和小于0时的根，并输出结果。从主函数输入a、b、c的值。 \n\n## 输入\na b c\n\n## 输出\nx1=? x2=?\n\n## 样例输入\n```\n4 1 1\n```\n\n\n## 样例输出\n```\nx1=-0.125+0.484i x2=-0.125-0.484i\n```\n', '', 1000, 131072, 0, '87a3101012aed9c5b19cebf5f05250b0', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:22', 0, 0, 0),
(84, '1073', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.3', '写一个判断素数的函数，在主函数输入一个整数，输出是否是素数的消息。\n\n## 输入\n一个数\n\n## 输出\n如果是素数输出prime\n如果不是输出not prime\n\n## 样例输入\n```\n97\n```\n\n\n## 样例输出\n```\nprime\n```\n', '', 1000, 131072, 0, '7286f0ea3baf1b65a286440736668f61', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:24', 0, 0, 0),
(85, '1074', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.4', '写一个函数，使给定的一个二维数组（３×３）转置，即行列互换。\n\n## 输入\n一个3x3的矩阵\n\n## 输出\n转置后的矩阵\n\n## 样例输入\n```\n1 2 3\n4 5 6\n7 8 9\n\n```\n\n\n## 样例输出\n```\n1 4 7 \n2 5 8 \n3 6 9 \n\n```\n', '', 1000, 131072, 0, '31a30ce20f3bb1a8a6db5c48970720fe', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:26', 0, 0, 0),
(86, '1075', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.5', '写一函数，使输入的一个字符串按反序存放，在主函数中输入输出反序后的字符串。  \n\n## 输入\n一行字符\n\n## 输出\n逆序后的字符串\n\n## 样例输入\n```\n123456abcdef \n\n```\n\n\n## 样例输出\n```\nfedcba654321\n\n```\n', '', 1000, 131072, 0, 'e3cb61bbf066c89ec0051b7ac22cc505', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:29', 0, 0, 0),
(87, '1076', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.6', '写一函数，将两个字符串连接\n\n## 输入\n两行字符串\n\n## 输出\n链接后的字符串\n\n## 样例输入\n```\n123\nabc\n```\n\n\n## 样例输出\n```\n123abc\n```\n', '', 1000, 131072, 0, '5f02474b7c23c9ee88f63c04502a9ac0', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:31', 0, 0, 0),
(88, '1077', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.7', '写一函数，将两个字符串中的元音字母复制到另一个字符串，然后输出。\n\n## 输入\n一行字符串\n\n## 输出\n顺序输出其中的元音字母（aeiuo）\n\n## 样例输入\n```\nabcde\n```\n\n\n## 样例输出\n```\nae\n```\n', '', 1000, 131072, 0, '65263aebb66fe92b03f6114263411a2a', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:34', 0, 0, 0),
(89, '1078', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.8', '写一函数，输入一个四位数字，要求输出这四个数字字符，但每两个数字间空格。如输入1990，应输出\"1 9 9 0\"。\n\n## 输入\n一个四位数\n\n## 输出\n增加空格输出\n\n## 样例输入\n```\n1990\n```\n\n\n## 样例输出\n```\n1 9 9 0 \n```\n', '', 1000, 131072, 0, '60e6be9b6cac50167a06b5eebea4e059', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:36', 0, 0, 0),
(90, '1079', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题8.9', '编写一函数，由实参传来一个字符串，统计此字符串中字母、数字、空格和其它字符的个数，在主函数中输入字符串以及输出上述结果。\n\n只要结果，别输出什么提示信息。\n\n## 输入\n一行字符串\n\n## 输出\n统计数据，4个数字，空格分开。\n\n## 样例输入\n```\n!@#$%^QWERT    1234567\n```\n\n\n## 样例输出\n```\n5 7 4 6 \n```\n', '', 1000, 131072, 0, 'ab38dbf1a71c2c4a7b283aaa655d09ec', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:39', 0, 0, 0),
(91, '108', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[4.3.1]判断奇偶', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">编写程序，输入一个整数，打印出它是奇数还是偶数。</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试有一个整数m(0&lt;=m&lt;=1000)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;line-height:1.5;font-family:\'Microsoft YaHei\';\">如果m是奇数则输出“odd”，若是偶数输出“even”。</span>\r\n</p>\n\n## 样例输入\n\n```\n2\r\n\n```\n\n## 样例输出\n\n```\neven\r\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, 'f2f82cf90593f4ced7b6323bda2bdae3', 3, '2016-11-09 22:16:58', 3, '2016-11-09 22:16:58', 309, 401, 0),
(92, '1080', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.1', '定义一个带参的宏，使两个参数的值互换，并写出程序，输入两个数作为使用宏时的实参。输出已交换后的两个值。  \n\n## 输入\n两个数，空格隔开\n\n## 输出\n交换后的两个数，空格隔开\n\n## 样例输入\n```\n1 2\n```\n\n\n## 样例输出\n```\n2 1\n```\n', '', 1000, 131072, 0, 'e2c153f19c1cfec6c3c2ced0e392898b', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:42', 0, 0, 0),
(93, '1081', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.2', '输入两个整数，求他们相除的余数。用带参的宏来实现，编程序。  \n\n## 输入\na b两个数\n\n## 输出\na/b的余数\n\n## 样例输入\n```\n3 2\n```\n\n\n## 样例输出\n```\n1\n```\n', '', 1000, 131072, 0, '90b0d3b5190e2ccba21c4fb68bb49a47', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:45', 0, 0, 0),
(94, '1082', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.3', '三角形面积=SQRT(S*(S-a)*(S-b)*(S-c))\n其中S=(a+b+c)/2，a、b、c为三角形的三边。\n定义两个带参的宏，一个用来求area，\n另一个宏用来求S。\n\n写程序，在程序中用带实参的宏名来求面积area。 \n\n## 输入\na b c三角形的三条边,可以是小数。\n\n## 输出\n三角形面积，保留3位小数\n\n## 样例输入\n```\n3 4 5\n```\n\n\n## 样例输出\n```\n6.000\n```\n', '', 1000, 131072, 0, 'd229c2229ca63a42119679768cf41d23', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:48', 0, 0, 0),
(95, '1083', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.4', '给年份year，定义一个宏，以判别该年份是否闰年。提示：宏名可以定义为LEAP_YEAR，形参为y，既定义宏的形式为  \n#define LEAP_YEAR(y) （读者设计的字符串）\n\n## 输入\n一个年份\n\n## 输出\n根据是否闰年输出，是输出\"L\",否输出\"N\"\n\n## 样例输入\n```\n2000\n```\n\n\n## 样例输出\n```\nL\n```\n', '', 1000, 131072, 0, '46e6e4d6117904043dd4b1417c5baf23', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:50', 0, 0, 0),
(96, '1084', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.6', '请设计输出实数的格式，包括：⑴一行输出一个实数；⑵一行内输出两个实数；⑶一行内输出三个实数。实数用\"6.2f\"格式输出。\n\n## 输入\n一个实数，float范围\n\n## 输出\n输出3行，第一行打印一遍输入的数，第二行打印两遍，第三行打印三遍。\n第二行和第三行，用空格分隔同一行的数字。\n实数用\"6.2f\"格式输出。\n\n## 样例输入\n```\n0.618\n```\n\n\n## 样例输出\n```\n  0.62\n  0.62   0.62\n  0.62   0.62   0.62\n\n```\n', '', 1000, 131072, 0, '45ab3a356e795cc339222009ccaa8045', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:53', 0, 0, 0),
(97, '1085', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.8', '分别用函数和带参的宏，从三个数中找出最大的数。\n\n## 输入\n3个实数\n\n## 输出\n最大的数,输出两遍，先用函数，再用宏。\n保留3位小数。\n\n## 样例输入\n```\n1 2 3\n```\n\n\n## 样例输出\n```\n3.000\n3.000\n```\n', '', 1000, 131072, 0, 'ec568e9d2762a93f264d69741174b437', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:56', 0, 0, 0),
(98, '1086', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题9.10', '输入一行电报文字，将字母变成其下一字母（如’a’变成’b’……’z’变成’ａ’其它字符不变）。\n\n## 输入\n一行字符\n\n## 输出\n加密处理后的字符\n\n## 样例输入\n```\na b\n```\n\n\n## 样例输出\n```\nb c\n```\n', '', 1000, 131072, 0, 'c9ceb6efe1a074a3fced9ced07af0d10', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:56:58', 0, 0, 0),
(99, '1087', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.1', '输入三个整数，按由小到大的顺序输出。\n\n## 输入\n三个整数\n\n## 输出\n由小到大输出成一行，每个数字后面跟一个空格\n\n## 样例输入\n```\n2 3 1\n```\n\n\n## 样例输出\n```\n1 2 3 \n```\n', '', 1000, 131072, 0, '19c68da707172de5cc32e145a1d4f2e1', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:01', 0, 0, 0),
(100, '1088', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.2', '输入三个字符串，按由小到大的顺序输出 \n\n## 输入\n3行字符串\n\n## 输出\n按照从小到大输出成3行\n\n## 样例输入\n```\ncde\nafg\nabc\n\n```\n\n\n## 样例输出\n```\nabc\nafg\ncde\n```\n', '', 1000, 131072, 0, 'b6ddbdf7ccb2a644245ce9973189ebeb', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:03', 0, 0, 0),
(101, '1089', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.3', '输入10个整数，将其中最小的数与第一个数对换，把最大的数与最后一个数对换。写三个函数；\n①输入10个数；②进行处理；③输出10个数。\n\n## 输入\n10个整数\n\n## 输出\n整理后的十个数，每个数后跟一个空格（注意最后一个数后也有空格）\n\n## 样例输入\n```\n2 1 3 4 5 6 7 8 10 9\n```\n\n\n## 样例输出\n```\n1 2 3 4 5 6 7 8 9 10 \n```\n', '', 1000, 131072, 0, '52b0359a71eb09e90b1062bbdf35680d', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:06', 0, 0, 0),
(102, '109', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[4.3.2]函数求值', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">编写程序计算下面的函数，要求输入x的值，输出y的值。</span>\r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">$$ y = \\begin{cases} x &amp; -5 &lt; x &lt; 0\\\\ x - 1 &amp; x = 0\\\\ x + 1 &amp; 0 &lt; x &lt; 10 \\end{cases} $$</span>\r\n</p>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试有一个实数x(-5&lt;x&lt;10)。</span>\r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">输出函数值(保留三位小数)。</span>\r\n</p>\n\n## 样例输入\n\n```\n0\r\n\n```\n\n## 样例输出\n\n```\n-1.000\r\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, '41ac14ae6c1473757031f0a7d213ca8a', 3, '2016-11-09 22:17:52', 3, '2016-11-09 22:17:52', 269, 476, 0),
(103, '1090', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.4', '有n个整数，使前面各数顺序向后移m个位置，最后m个数变成前面m个数，见图。写一函数：实现以上功能，在主函数中输入n个数和输出调整后的n个数。\n\n## 输入\n输入数据的个数n\nn个整数\n移动的位置m\n\n## 输出\n移动后的n个数\n\n## 样例输入\n```\n10\n1 2 3 4 5 6 7 8 9 10\n2\n```\n\n\n## 样例输出\n```\n9 10 1 2 3 4 5 6 7 8 \n```\n', '', 1000, 131072, 0, 'a97bb6b87682c0c9021e55496661e03f', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:08', 0, 0, 0),
(104, '1091', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.5', '有n人围成一圈，顺序排号。从第1个人开始报数（从1到3报数），凡报到3的人退出圈子，问最后留下的是原来的第几号的那位。\n\n## 输入\n初始人数n\n\n## 输出\n最后一人的初始编号\n\n## 样例输入\n```\n3\n```\n\n\n## 样例输出\n```\n2\n```\n', '', 1000, 131072, 0, 'fa494bb6caad2eae76e6c2487d1ca991', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:11', 0, 0, 0),
(105, '1092', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题10.7', '有一字符串，包含n个字符。写一函数，将此字符串中从第m个字符开始的全部字符复制成为另一个字符串。\n\n## 输入\n数字n\n一行字符串\n数字m\n\n## 输出\n从m开始的子串\n\n## 样例输入\n```\n6\nabcdef\n3\n```\n\n\n## 样例输出\n```\ncdef\n```\n', '', 1000, 131072, 0, '15e8fa191d0735888d858e309ac1775f', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:13', 0, 0, 0),
(106, '1093', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题11.1', '定义一个结构体变量（包括年、月、日）。计算该日在本年中是第几天，注意闰年问题。  \n\n\n## 输入\n年月日\n\n## 输出\n当年第几天\n\n## 样例输入\n```\n2000 12 31\n```\n\n\n## 样例输出\n```\n366\n```\n', '', 1000, 131072, 0, 'be3b2b6719709d445843730568eeeb78', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:16', 0, 0, 0),
(107, '1094', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题11.3', '现有有N个学生的数据记录，每个记录包括学号、姓名、三科成绩。\n编写一个函数input,用来输入一个学生的数据记录。\n编写一个函数print,打印一个学生的数据记录。\n在主函数调用这两个函数，读取N条记录输入，再按要求输出。\nN<100\n\n## 输入\n学生数量N占一行\n每个学生的学号、姓名、三科成绩占一行，空格分开。\n\n## 输出\n每个学生的学号、姓名、三科成绩占一行，逗号分开。\n\n## 样例输入\n```\n2\na100 zhblue 70 80 90\nb200 newsclan 90 85 75\n\n```\n\n\n## 样例输出\n```\na100,zhblue,70,80,90\nb200,newsclan,90,85,75\n\n```\n', '', 1000, 131072, 0, '9f877f07bbb05c96ba860c0cec87f2a5', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:19', 0, 0, 0),
(108, '1095', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题11.5', '有N个学生，每个学生的数据包括学号、姓名、3门课的成绩，从键盘输入N个学生的数据，要求打印出3门课的总平均成绩，以及最高分的学生的数据（包括学号、姓名、3门课成绩）\n\n## 输入\n学生数量N占一行每个学生的学号、姓名、三科成绩占一行，空格分开。\n\n## 输出\n各门课的平均成绩\n最高分的学生的数据（包括学号、姓名、3门课成绩）\n\n## 样例输入\n```\n2\n1 blue 90 80 70\nb clan 80 70 60\n```\n\n\n## 样例输出\n```\n85 75 65\n1 blue 90 80 70\n```\n', '', 1000, 131072, 0, 'a958c918944dd67c64ff83b3db88cd54', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:22', 0, 0, 0),
(109, '1096', 4, NULL, NULL, NULL, NULL, 'C语言程序设计教程（第三版）课后习题11.8', '已有a、b两个链表，每个链表中的结点包括学好、成绩。要求把两个链表合并，按学号升序排列。  \n\n## 输入\n第一行，a、b两个链表元素的数量N、M,用空格隔开。\n接下来N行是a的数据\n然后M行是b的数据\n每行数据由学号和成绩两部分组成\n\n\n\n## 输出\n按照学号升序排列的数据\n\n## 样例输入\n```\n2 3\n5 100\n6 89\n3 82\n4 95\n2 10\n```\n\n\n## 样例输出\n```\n2 10\n3 82\n4 95\n5 100\n6 89\n```\n', '', 1000, 131072, 0, '72235ccb5cbcde59771c2ab8266923ca', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:25', 0, 0, 0),
(110, '1097', 4, NULL, NULL, NULL, NULL, '词组缩写', '<p>定义：一个词组中每个单词的首字母的大写组合称为该词组的缩写。 比如，C语言里常用的EOF就是end of file的缩写。</p>\n\n## 输入\n<p>输入的第一行是一个整数T，表示一共有T组测试数据； 接下来有T行，每组测试数据占一行，每行有一个词组，每个词组由一个或多个单词组成；每组的单词个数不超过10个，每个单词有一个或多个大写或小写字母组成； 单词长度不超过10，由一个或多个空格分隔这些单词。</p>\n\n## 输出\n<p>请为每组测试数据输出规定的缩写，每组输出占一行。</p>\n\n## 样例输入\n```\n1\nend of file \n```\n\n\n## 样例输出\n```\nEOF\n```\n', '绍兴托普信', 1000, 131072, 0, 'b9c424ae6893cdb16db1ae9c9613646d', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:28', 0, 0, 0),
(111, '1098', 4, NULL, NULL, NULL, NULL, '18岁', '<p>小明的18岁生日就要到了，他当然很开心，可是他突然想到一个问题，是不是每个人从出生开始，到达18岁生日时所经过的天数都是一样的呢？似乎并不全都是这样，所以他想请你帮忙计算一下他和他的几个朋友从出生到达18岁生日所经过的总天数，让他好来比较一下。<br />\n</p>\n\n## 输入\n<p>输入的第一行是一个数T，后面T行每行有一个日期，格式是YYYY-MM-DD。如我的生日是1988-03-07。</p>\n\n## 输出\n<p>T行，每行一个数，表示此人从出生到18岁生日所经过的天数。如果这个人没有18岁生日，就输出-1。<br />\n</p>\n\n## 样例输入\n```\n1\n1988-03-07\n```\n\n\n## 样例输出\n```\n6574\n```\n', '', 1000, 131072, 0, 'b5440bfe256626dbae99bba6d236514e', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:30', 0, 0, 0),
(112, '1099', 4, NULL, NULL, NULL, NULL, 'HDB3码', '<p>villa在学习通信原理课程，其中有一个知识点，即为HDB3编码，她怎么也弄不明白，作为ACMer 你能帮帮她吗？<br />\nAMI码就是把单极性脉冲序列中相邻的&ldquo;1&rdquo;码(即正脉冲)变为极性交替的正、负脉冲。将&ldquo;0&rdquo;码保持不变，把&ldquo;1&rdquo;<br />\n码变为+1、-1交替的脉冲。如：<br />\n&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 消息码： 1 0 0 0 0&nbsp; 1 0 0 0 0&nbsp; 1&nbsp; 1 0 0 0 0&nbsp; 1&nbsp; 1&nbsp; <br />\n&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp; AMI码：-1 0 0 0 0 +1 0 0 0 0 -1 +1 0 0 0 0 -1 +1 <br />\nHDB3码（3nd Order High Density Bipolar）的全称是三阶高密度双极性码，它是AMI码的一种改进型，改进目的是为<br />\n了保持AMI码的优点而克服其缺点，使连&ldquo;0&rdquo;个数不超过3个。其编码规则：<br />\n（1）检查消息码中连&ldquo;0&rdquo;的个数。当连&ldquo;0&rdquo;数目小于等于3时，HDB3码与AMI码一样（&ldquo;1&rdquo;交替的变换为&ldquo;+1&rdquo;和<br />\n&ldquo;-1&rdquo;，&ldquo;0&rdquo;保持不变）。&nbsp; <br />\n（2）当连&ldquo;0&rdquo;数目超过3时，将每4个连&ldquo;0&rdquo;化作一小节，定义为&ldquo;B00V&quot;称为破坏节，其中V称为破坏脉冲，而B称<br />\n为调节脉冲；&nbsp; <br />\n（3）V与前一个相邻的非&ldquo;0&rdquo;脉冲的极性相同（这破坏了极性交替的规则，所以V称破坏脉冲），并且要求相邻的V码<br />\n之间极性必须交替。V的取值为&ldquo;+1&rdquo;或&ldquo;-1&rdquo;；&nbsp; <br />\n（4）B的取值可选0、+1或-1，（当B取+1的时候用+B表示，B取-1的时候用-B表示，B取0时用0表示）以使V同时满足（3）<br />\n中的两个要求；&nbsp; <br />\n（5）除去V码以外的非零码（即1码和B码）的极性交替。<br />\n我们规定编码后第一个非零码元为-1，第一个B的值为0。且给定的消息码中第一个码元为1.<br />\n例如： <br />\n&nbsp;&nbsp;&nbsp; 消息码： 1 000 0&nbsp; 1 000 0&nbsp; 1 1 000 0&nbsp; 000 0&nbsp; 1 1 <br />\n&nbsp;&nbsp;&nbsp; AMI码： -1 000 0 +1 000 0 -1+1 000 0&nbsp; 000 0 -1+1 <br />\n&nbsp;&nbsp;&nbsp; HDB3码：-1 000-V +1 000+V -1+1-B00-V +B00+V -1+1<br />\n（样例解释：根据规则3，第三个V极性为负，而其前面的非零码1的极性为正，所以添加-B调节，使其同时满足规则3的两个条件。）<br />\n</p>\n\n## 输入\n<p>输入包括多组测试数据，其中第一行为一个整数n表示有n组测试数据（0&lt;n&lt;10000），接下来包括n行数据。每一行的第一个数据整数<br />\nm(0&lt;=m&lt;1000)表示的是码元个数，接下来为m个&ldquo;0&rdquo;或&ldquo;1&rdquo;的整数。</p>\n\n## 输出\n<p>每组测试数据输出一行编码后的HDB3码。</p>\n\n## 样例输入\n```\n1\n22 1 0 0 0 0 1 0 0 0 0 1 1 0 0 0 0 0 0 0 0 1 1\n\n```\n\n\n## 样例输出\n```\n-1000-V+1000+V-1+1-B00-V+B00+V-1+1\n\n```\n', '', 1000, 131072, 0, 'e959ab43825e46184ca79d89bc79cf7d', 3, '2025-05-28 22:04:30', 3, '2025-05-29 17:57:52', 0, 0, 0),
(113, '11', 2, NULL, NULL, NULL, NULL, 'QAQ的序列价值', '## 问题描述\n\nQAQ有一个序列，元素个数有$ N $个。\n\n他认为一个序列的价值的是：该序列中不同元素之和。\n\n比如说：序列$ (1, 1, 2, 2) $价值为$ 3 $。\n\n现在QAQ想知道所有子序列的价值之和。\n\n## 输入\n\n第一行输入一个整数$T$，代表有$T$组测试数据。\n\n每组数据占两行，第一行输入一个整数$N$，代表序列元素个数。\n\n接下来一行输入$N$个整数$a[]$。\n\n注：$ 1 \\leq T  \\leq 10000 $，$ 1  \\leq N  \\leq 50 $，$ 1  \\leq a[]  \\leq 10 $。\n\n## 输出\n\n对每组测试数据，输出一个整数代表所有子序列价值之和。\n\n结果很大，请对$ (10^9 + 7) $取余。\n\n## 样例输入\n\n```\n2\n3\n1 1 1\n4\n10 10 10 8\n```\n\n## 样例输出\n\n```\n7\n204\n```\n\n## 提示\n\n对于第二组测试数据一共有$ 15 $个子序列：\n$(10)$、$(10)$、$(10)$、$(8)$、$(10, 10)$、$(10, 10)$、$(10, 10)$、$(10, 8)$、$(10, 8)$、$(10, 8)$、$(10, 10, 8)$、$(10, 10, 8)$、$(10, 10, 8)$、$(10, 10, 10)$、$(10, 10, 10, 8)$\n\n价值之和为$ 204 $\n\n', 'CZY', 3000, 131072, 0, 'c960ec8d233a1f3bcb6c90cc27d066ac', 3, '2016-10-26 21:17:27', 3, '2025-06-21 18:50:08', 25, 81, 0),
(114, '110', 3, NULL, NULL, NULL, NULL, '【C语言程序设计】[4.3.3]改写语句', '## 问题描述\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">当输入一个整数a，请输出按照如下代码处理之后的m的值。</span> \r\n</p>\r\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;\"> </span> \r\n</p>\r\n<pre class=\"prettyprint lang-cpp\">	if(a&lt;30) m=1;\r\n	else if(a&lt;40) m=2;\r\n	else if(a&lt;50) m=3;\r\n	else if(a&lt;60) m=4;\r\n	else m=5;</pre>\n\n## 输入\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">每次测试有一个整数a(a&lt;1000)。</span> \r\n</p>\n\n## 输出\n\n<p style=\"text-indent:2em;\">\r\n	<span style=\"font-size:18px;font-family:\'Microsoft YaHei\';\">请输出经过处理后m的值。</span> \r\n</p>\n\n## 样例输入\n\n```\n100\r\n\n```\n\n## 样例输出\n\n```\n5\r\n\n```', '贾宗璞许合利主编', 1000, 131072, 0, '559cbba4f9940fc477d551eb5f211daa', 3, '2016-11-09 22:18:20', 3, '2016-11-09 22:18:20', 272, 321, 0),
(115, '1100', 4, NULL, NULL, NULL, NULL, '反射', '<p><br />\n我们都知道人眼对自然界中的景物有不同的彩色感觉是因为日光（光源）包含有各种波长的可见光，<br />\n同时景物又具有不同的反射（或透射）特性的缘故。例如：西红柿具有反射红色光谱成分的特性，<br />\n在太阳光照射下其中的红色成分（吸收了其他光谱成分），所以人眼看到的西红柿是红色的。如果<br />\n把西红柿放到蓝光下，照射到西红柿上的蓝光被它吸收了，人眼看到的将是黑色的西红柿。</p>\n\n## 输入\n<p>我们在每组测试数据的第一行有一个字符，表示某个景物能够反射的颜色，第二行输入的是一行<br />\n字符串（长度不超过1000），表示一组照射到该景物上的光源（长度不超过26）。问，是否有光源<br />\n照射到景物上能够显示出景物的颜色，如果有则将每个光源按行输出，没有则输出&ldquo;No!&rdquo;。<br />\n有多组测试数据，每组测试数据的输出结果用一行空行分割开来。</p>\n\n## 样例输入\n```\nr\nogr bg abc r\ne\noga bw\n```\n\n\n## 样例输出\n```\nogr\nr\n\nNo!\n```\n', '', 1000, 131072, 0, '61360cf32608f1c67e335e1bb6e8e295', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:57:56', 0, 0, 0),
(116, '1101', 4, NULL, NULL, NULL, NULL, '求日龄', '<p>年龄是指一个人出生后以年为长度计算的时间整数值。</p>\n<p>同理，日龄指一个人出生后以日为长度计算的时间整数值。</p>\n<p>请求出给定日期出生的人，到2013年3月9日的日龄吧。</p>\n\n## 输入\n<p>一个日期，表示这个人的生日</p>\n<p>输入日期一定在2013-3-9之前</p>\n\n## 输出\n<p>日龄</p>\n\n## 样例输入\n```\n2013-3-8\n```\n\n\n## 样例输出\n```\n1\n```\n', '', 1000, 131072, 0, 'e44f90a6402f6592d947aa5ec66e4746', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:01', 0, 0, 0),
(117, '1102', 4, NULL, NULL, NULL, NULL, '奇偶间谍', '<p>奇数和偶数正在交战，在奇数的阵营里有偶数出现就被认为是间谍，反过来也一样。</p>\n<p>输入包含10行，请将其中的间谍数（与其他数奇偶性不同的那一个）找出来并输出。</p>\n<p>注意，间谍可能是奇数也可能是偶数</p>\n\n## 输入\n<p>10行，每行一个数字</p>\n\n## 输出\n<p>1行，与其他数不同的那一个</p>\n\n## 样例输入\n```\n1\n3\n5\n7\n9\n11\n13\n14\n17\n19\n```\n\n\n## 样例输出\n```\n14\n\n```\n', '', 1000, 131072, 0, '01ab8b99c5e50ee7a17c7002e5fa3ebc', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:04', 0, 0, 0),
(118, '1103', 4, NULL, NULL, NULL, NULL, '笨鸟先飞', '<p>多多是一只小菜鸟，都说笨鸟先飞，多多也想来个菜鸟先飞。于是它从0点出发，一开始的飞行速度为1m/s，每过一个单位时间多多的飞行速度比上一个单位时间的飞行速度快2m/s，问n（0&lt;n&lt;10^5）个单位时间之后多多飞了多远?<br />\n<br />\n</p>\n\n## 输入\n<p><br />\n先输入一个整数T表示有几组数据。每组数据输入一个n，表示多多飞行的时间。<br />\n</p>\n\n## 输出\n<p>输出多多飞行了多远，因为数字很大，所以对10000取模。<br />\n<br />\n</p>\n\n## 样例输入\n```\n2\n1\n2\n```\n\n\n## 样例输出\n```\n1\n4\n```\n', '', 1000, 131072, 0, 'e0e4fea49c6894448099547faee92085', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:07', 0, 0, 0),
(119, '1104', 4, NULL, NULL, NULL, NULL, '关系推断', '<p>给你一些已经确定的元素之间的关系，请你判断是否能从这些元素关系中推断出其他的元素关系。<br />\n</p>\n\n## 输入\n<p>输入的第一行是一个整数N，表示测试数据的组数。<br />\n每组输入首先是一个正整数m（m&lt;=100），表示给定元素关系的个数。<br />\n接下来m行，每行一个元素关系，格式为：<br />\n元素1&lt;元素2&nbsp; 或者&nbsp; 元素1&gt;元素2<br />\n元素用一个大写字母表示，输入中不会包含冲突的关系。<br />\n</p>\n\n## 输出\n<p>对于每组输入，第一行输出&ldquo;Case d:&rdquo;，d是测试数据的序号，从1开始。<br />\n接下来输出所有推断出的新的元素关系，按照字典序从小到大排序，格式为：<br />\n元素1&lt;元素2<br />\n每个元素关系占一行，输入中给定的元素关系不要输出。<br />\n如果没有新的元素关系推断出来，则输出NONE。<br />\n</p>\n\n## 样例输入\n```\n2\n3\nA<B\nC>B\nC<D\n2\nA<B\nC<D\n```\n\n\n## 样例输出\n```\nCase 1:\nA<C\nA<D\nB<D\nCase 2:\nNONE\n```\n', '', 1000, 131072, 0, '10a92a70f2e18e69a99b76f2b1244049', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:09', 0, 0, 0),
(120, '1105', 4, NULL, NULL, NULL, NULL, '时钟', '<p>有一个标准的12小时时钟，它有一个时针，一个分针。现问你，在给定的两个时刻之间分针与时针相遇几次？<br />\n<br />\n</p>\n\n## 输入\n<p>输入包含多组测试数据。每组输入包含4个整数，前两个数字分别表示起始时刻的小时和分，后两个数字分别表示结束时刻的小时和分。<br />\n小时数在[1，12]内，分钟数在[0，59]内。<br />\n注意：<br />\n1）输入中的起始和结束时刻均不会出现时针和分针恰好相遇的情况，例如12点0分。<br />\n2）输入中不会出现起始时刻和结束时刻相同的情况。<br />\n3）在时针从起始时刻到结束时刻运转的过程中，时针转过的角度一定小于360度。<br />\n4）在时针从起始时刻到结束时刻运转的过程中，时针有可能越过表盘上12点钟的刻度。如果越过了，说明起始时刻和结束时刻中一个是A.M.，一个是P.M.。如果没越过，说明起始时刻和结束时刻都是A.M.或都是P.M.。<br />\n<br />\n</p>\n\n## 输出\n<p>输出的第一行为&ldquo;Initial time&nbsp; Final time&nbsp; Passes&rdquo;，具体的输出格式请参照输出样例。可用鼠标选中来看出空格数等细节。<br />\n</p>\n\n## 样例输入\n```\n12 50  1  2\n 3  8  3 20\n 2 45 11  0\n11  0  3 20\n 1  2 12 50\n 3 20  3  8\n\n```\n\n\n## 样例输出\n```\nInitial time  Final time  Passes\n       12:50       01:02       0\n       03:08       03:20       1\n       02:45       11:00       8\n       11:00       03:20       4\n       01:02       12:50      11\n       03:20       03:08      10\n```\n', '', 1000, 131072, 0, 'adefcc0e18ea2890963bf7f90470dc6b', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:12', 0, 0, 0),
(121, '1106', 4, NULL, NULL, NULL, NULL, '统计方案', '<p>在一无限大的二维平面中，我们做如下假设：<br />\n1、每次只能移动一格；<br />\n2、不能向后走（假设你的目的地是&ldquo;向上&rdquo;，那么你可以向左走，可以向右走，也可以向上走，但是不可以向下走）；<br />\n3、走过的格子立即塌陷无法再走第二次。<br />\n求走n步不同的方案数（2种走法只要有一步不一样，即被认为是不同的方案）。<br />\n</p>\n\n## 输入\n<p><br />\n<br />\n首先给出一个正整数C，表示有C组测试数据。<br />\n接下来的C行，每行包含一个整数n（n&lt;=20），表示要走n步。<br />\n</p>\n\n## 输出\n<p><br />\n请编程输出走n步的不同方案总数；<br />\n每组的输出占一行。<br />\n</p>\n\n## 样例输入\n```\n2\n1\n2\n```\n\n\n## 样例输出\n```\n3\n7\n```\n', '', 1000, 131072, 0, 'aad779704f3a8d8baedae3567b24cac4', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:15', 0, 0, 0),
(122, '1107', 4, NULL, NULL, NULL, NULL, 'Problem A Quadratic Sum', '<p>(Standard Input / Standard Output)<br />\nGiven a sequence of integers, you should output the quadratic sum of them. For example, the quadratic sum of sequence 3, -1, 8 is 32 + (-1)2 + 82.</p>\n\n## 输入\n<p>The first line of input contains a number representing the number of test cases.<br />\nThe first line of each test case contains an integer N (1 &lt;= N &lt;= 100), which is the number of integers in the sequence. The next line contains N integers which range are [-100, 100].<br />\n</p>\n\n## 输出\n<p>For each test case, output the quadratic sum of the given sequence in a one line.</p>\n\n## 样例输入\n```\n2\n3\n3 -1 8\n4\n2 4 6 8\n\n```\n\n\n## 样例输出\n```\n74\n120\n\n```\n', '2011 ACM-ICPC Zhuhai Metropolitan\nProgramming Contest\n', 1000, 131072, 0, '196311bc65f93f901c3eec8cc213f3bb', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:17', 0, 0, 0),
(123, '1108', 4, NULL, NULL, NULL, NULL, 'Problem B Manhattan', '<p>(Standard Input / Standard Output)<br />\nAs a new comer of Manhattan, you want to know the minimum manhattan distance among all the skyscrapers.<br />\nThere are N skyscrapers numbered from 1 to N, and calculated their cartesian coordinates (Xi, Yi). The manhattan distance of two distinct skyscrapers with cartesian coordinates marked as (Xa, Ya) and (Xb, Yb) is | Xa - Xb | + | Ya - Yb | (Notice that | x | means the absolute value of x). Please calculate the the minimum manhattan distance among all legal pairs of skyscrapers.<br />\n</p>\n\n## 输入\n<p>The first line contains the number of test cases T (T &lt;= 25). Following lines are the scenarios of each test case.<br />\nThe first line of each test case contains one integer N (2&lt;=N&lt;=100). The following N lines describe the coordinates of the skyscrapers. Each of these lines will contain two integers Xi and Yi (-100000 &lt;= Xi, Yi &lt;= -100000) representing the coordinates of the corresponding skyscrapers. The coordinates of the skyscrapers will not coincide with each other.<br />\n</p>\n\n## 输出\n<p>For each test case, please output one integer representing the minimum manhattan distance.</p>\n\n## 样例输入\n```\n1\n4\n-1 0\n0 -3\n2 0\n2 2\n\n```\n\n\n## 样例输出\n```\n2\n```\n', '2011 ACM-ICPC Zhuhai Metropolitan\nProgramming Contest\n', 1000, 131072, 0, '892ba5c78b18ba99f3241f83ecca4ce7', 3, '2025-05-28 22:04:31', 3, '2025-05-29 17:58:20', 0, 0, 0);

-- --------------------------------------------------------

--
-- 表的结构 `problem_daily`
--

CREATE TABLE `problem_daily` (
  `key` char(10) NOT NULL,
  `problem_id` int(11) NOT NULL,
  `solution` text NOT NULL,
  `code` text NOT NULL,
  `inserter` int(11) DEFAULT NULL,
  `modifier` int(11) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `problem_member`
--

CREATE TABLE `problem_member` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='问题访问权限';

-- --------------------------------------------------------

--
-- 表的结构 `problem_member_auth`
--

CREATE TABLE `problem_member_auth` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='问题管理权限';

-- --------------------------------------------------------

--
-- 表的结构 `problem_tag`
--

CREATE TABLE `problem_tag` (
  `id` int(11) NOT NULL,
  `tag_id` int(11) NOT NULL,
  `index` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `tag`
--

CREATE TABLE `tag` (
  `id` int(11) NOT NULL,
  `name` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `user`
--

CREATE TABLE `user` (
  `id` int(11) NOT NULL,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `nickname` varchar(20) NOT NULL COMMENT '昵称',
  `real_name` varchar(20) DEFAULT NULL COMMENT '真实名称',
  `password` char(32) DEFAULT NULL,
  `email` varchar(20) DEFAULT NULL,
  `number` varchar(20) DEFAULT NULL COMMENT '身份标识',
  `qq` varchar(15) DEFAULT NULL COMMENT 'QQ',
  `vjudge_id` varchar(15) DEFAULT NULL COMMENT 'VjudgeId',
  `github` varchar(15) DEFAULT NULL COMMENT 'Github',
  `codeforces` varchar(15) DEFAULT NULL,
  `slogan` varchar(30) DEFAULT NULL COMMENT '签名',
  `organization` varchar(20) DEFAULT NULL COMMENT '组织',
  `insert_time` datetime DEFAULT NULL,
  `accept` int(11) DEFAULT NULL,
  `attempt` int(11) DEFAULT NULL,
  `check_in_count` int(11) DEFAULT NULL COMMENT '签到次数',
  `gender` tinyint(1) DEFAULT NULL COMMENT '性别'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- 表的结构 `user_role`
--

CREATE TABLE `user_role` (
  `id` int(11) DEFAULT NULL,
  `role_id` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- 转储表的索引
--

--
-- 表的索引 `collection`
--
ALTER TABLE `collection`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `collection_member`
--
ALTER TABLE `collection_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `collection_problem`
--
ALTER TABLE `collection_problem`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `group_id` (`id`,`problem_id`);

--
-- 表的索引 `contest`
--
ALTER TABLE `contest`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `contest_language`
--
ALTER TABLE `contest_language`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `contest_member`
--
ALTER TABLE `contest_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_auth`
--
ALTER TABLE `contest_member_auth`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_ignore`
--
ALTER TABLE `contest_member_ignore`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_volunteer`
--
ALTER TABLE `contest_member_volunteer`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_problem`
--
ALTER TABLE `contest_problem`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `group_id` (`id`,`problem_id`),
  ADD UNIQUE KEY `group_index` (`id`,`index`);

--
-- 表的索引 `discuss`
--
ALTER TABLE `discuss`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `discuss_pk_2` (`id`);

--
-- 表的索引 `discuss_comment`
--
ALTER TABLE `discuss_comment`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `discuss_pk` (`id`);

--
-- 表的索引 `discuss_tag`
--
ALTER TABLE `discuss_tag`
  ADD PRIMARY KEY (`id`,`tag_id`);

--
-- 表的索引 `judger`
--
ALTER TABLE `judger`
  ADD PRIMARY KEY (`key`);

--
-- 表的索引 `judge_job`
--
ALTER TABLE `judge_job`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `judge_job_compile`
--
ALTER TABLE `judge_job_compile`
  ADD UNIQUE KEY `judge_job_compile_pk` (`id`);

--
-- 表的索引 `problem`
--
ALTER TABLE `problem`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `problem_pk` (`key`),
  ADD UNIQUE KEY `problem_origin` (`origin_oj`,`origin_id`);

--
-- 表的索引 `problem_daily`
--
ALTER TABLE `problem_daily`
  ADD PRIMARY KEY (`key`),
  ADD UNIQUE KEY `problem_daily_pk_2` (`problem_id`);

--
-- 表的索引 `problem_member`
--
ALTER TABLE `problem_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `problem_member_auth`
--
ALTER TABLE `problem_member_auth`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `problem_tag`
--
ALTER TABLE `problem_tag`
  ADD PRIMARY KEY (`id`,`tag_id`);

--
-- 表的索引 `tag`
--
ALTER TABLE `tag`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_username` (`username`);

--
-- 表的索引 `user_role`
--
ALTER TABLE `user_role`
  ADD UNIQUE KEY `group_id` (`id`,`role_id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `collection`
--
ALTER TABLE `collection`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `contest`
--
ALTER TABLE `contest`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `discuss`
--
ALTER TABLE `discuss`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `discuss_comment`
--
ALTER TABLE `discuss_comment`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `judge_job`
--
ALTER TABLE `judge_job`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `problem`
--
ALTER TABLE `problem`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '问题ID', AUTO_INCREMENT=124;

--
-- 使用表AUTO_INCREMENT `tag`
--
ALTER TABLE `tag`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `user`
--
ALTER TABLE `user`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
