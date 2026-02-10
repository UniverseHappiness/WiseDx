# 数据库表
运行命令： docker exec -it WeKnora-postgres-dev psql -U postgres -d WeKnora
常用 psql 命令：
\dt：查看所有表。
\d table_name：查看具体表的结构。
SELECT * FROM users LIMIT 10;：执行 SQL 查询语句（注意末尾的分号）。
\q：退出数据库命令行。
\l：列出所有数据库。

### tenants
```
                                           Table "public.tenants"
       Column        |           Type           | Collation | Nullable |               Default               
---------------------+--------------------------+-----------+----------+-------------------------------------
 id                  | integer                  |           | not null | nextval('tenants_id_seq'::regclass)
 name                | character varying(255)   |           | not null | 
 description         | text                     |           |          | 
 api_key             | character varying(64)    |           | not null | 
 retriever_engines   | jsonb                    |           | not null | '[]'::jsonb
 status              | character varying(50)    |           |          | 'active'::character varying
 business            | character varying(255)   |           | not null | 
 storage_quota       | bigint                   |           | not null | '10737418240'::bigint
 storage_used        | bigint                   |           | not null | 0
 agent_config        | jsonb                    |           |          | 
 created_at          | timestamp with time zone |           |          | CURRENT_TIMESTAMP
 updated_at          | timestamp with time zone |           |          | CURRENT_TIMESTAMP
 deleted_at          | timestamp with time zone |           |          | 
 context_config      | jsonb                    |           |          | 
 conversation_config | jsonb                    |           |          | 
 web_search_config   | jsonb                    |           |          | 
Indexes:
    "tenants_pkey" PRIMARY KEY, btree (id)
    "idx_tenants_agent_config" gin (agent_config)
    "idx_tenants_api_key" btree (api_key)
    "idx_tenants_status" btree (status)
Referenced by:
    TABLE "users" CONSTRAINT "fk_users_tenant" FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL

```


举例：
---
id                  | 10001
name                | user2's Workspace
description         | Default workspace
api_key             | sk-XofsbvXmTB0_Mah6t9ybfh0x_T_9_8da-ni2zpQ6U_AkwhHY
retriever_engines   | {"engines": []}
status              | active
business            | 
storage_quota       | 10737418240
storage_used        | 0
agent_config        | 
created_at          | 2026-02-08 12:07:45.419976+00
updated_at          | 2026-02-08 12:07:45.421905+00
deleted_at          | 
context_config      | null
conversation_config | 
web_search_config   | 



id                  | 10000
name                | user1's Workspace
description         | Default workspace
api_key             | sk-dsve2JfFpBfJh6k3R5zaVmRAOgK1KTifwnI_CwKLdgopz5YE
retriever_engines   | {"engines": []}
status              | active
business            | 
storage_quota       | 10737418240
storage_used        | 0
agent_config        | 
created_at          | 2026-02-08 11:40:19.164076+00
updated_at          | 2026-02-08 12:53:16.724354+00
deleted_at          | 
context_config      | {"max_tokens": 0, "summarize_threshold": 0, "compression_strategy": "", "recent_message_cou
nt": 0}
conversation_config | {"prompt": "你是一个专业的智能信息检索助手，名为WeKnora。你犹如专业的高级秘书，依据检索到的
信息回答用户问题，不能利用任何先验知识。\n当用户提出问题时，助手会基于特定的信息进行解答。助手首先在心中思考推理
过程，然后向用户提供答案。\n## 回答问题规则\n- 仅根据检索到的信息中的事实进行回复，不得运用任何先验知识，保持回应
的客观性和准确性。\n- 复杂问题和答案的按Markdown分结构展示，总述部分不需要拆分\n- 如果是比较简单的答案，不需要把
最终答案拆分的过于细碎\n- 结果中使用的图片地址必须来自于检索到的信息，不得虚构\n- 检查结果中的文字和图片是否来自
于检索到的信息，如果扩展了不在检索到的信息中的内容，必须进行修改，直到得到最终答案\n- 如果用户问题无法回答，必须
如实告知用户，并给出合理的建议。\n\n## 输出限制\n- 以Markdown图文格式输出你的最终结果\n- 输出内容要保证简短且全面
，条理清晰，信息明确，不重复。\n", "max_rounds": 5, "temperature": 0.3, "rerank_top_k": 5, "enable_rewrite": true
, "embedding_top_k": 10, "fallback_prompt": "你是一个专业、友好的AI助手。请根据你的知识直接回答用户的问题。\n\n##
 回复要求\n- 直接回答用户的问题\n- 简洁清晰，言之有物\n- 如果涉及实时数据或个人隐私信息，诚实说明无法获取\n- 使用
礼貌、专业的语气\n\n## 用户的问题是:\n{{query}}\n", "rerank_model_id": "", "context_template": "{{query}}", "rera
nk_threshold": 0.5, "summary_model_id": "d1a3eb76-0d96-429d-99b2-cce9398b36a2", "vector_threshold": 0.5, "fallbac
k_response": "抱歉，我无法回答这个问题。", "fallback_strategy": "model", "keyword_threshold": 0.3, "rewrite_promp
t_user": "## 历史对话背景\n{{conversation}}\n\n## 需要改写的用户问题\n{{query}}\n\n## 改写后的问题\n", "max_compl
etion_tokens": 2048, "rewrite_prompt_system": "你是一个专注于指代消解和省略补全的智能助手，你的任务是根据历史对话
上下文，清晰识别用户问题中的代词并替换为明确的主语，同时补全省略的关键信息。\n\n## 改写目标\n请根据历史对话，对当
前用户问题进行改写，目标是：\n- 进行指代消解，将\"它\"、\"这个\"、\"那个\"、\"他\"、\"她\"、\"它们\"、\"他们\"、\
"她们\"等代词替换为明确的主语\n- 补全省略的关键信息，确保问题语义完整\n- 保持问题的原始含义和表达方式不变\n- 改写
后必须也是一个问题\n- 改写后的问题字数控制在30字以内\n- 仅输出改写后的问题，不要输出任何解释，更不要尝试回答该问
题，后面有其他助手回去解答此问题\n\n## Few-shot示例\n\n示例1:\n历史对话:\n用户: 微信支付有哪些功能？\n助手: 微信
支付的主要功能包括转账、付款码、收款、信用卡还款等多种支付服务。\n\n用户问题: 它的安全性\n改写后: 微信支付的安全
性\n\n示例2:\n历史对话:\n用户: 苹果手机电池不耐用怎么办？\n助手: 您可以通过降低屏幕亮度、关闭后台应用和定期更新系
统来延长电池寿命。\n\n用户问题: 这样会影响使用体验吗？\n改写后: 降低屏幕亮度和关闭后台应用是否影响使用体验\n\n示
例3:\n历史对话:\n用户: 如何制作红烧肉？\n助手: 红烧肉的制作需要先将肉块焯水，然后加入酱油、糖等调料慢炖。\n\n用户
问题: 需要炖多久？\n改写后: 红烧肉需要炖多久\n\n示例4:\n历史对话:\n用户: 北京到上海的高铁票价是多少？\n助手: 北京
到上海的高铁票价根据车次和座位类型不同，二等座约为553元，一等座约为933元。\n\n用户问题: 时间呢？\n改写后: 北京到
上海的高铁时长\n\n示例5:\n历史对话:\n用户: 如何注册微信账号？\n助手: 注册微信账号需要下载微信APP，输入手机号，接
收验证码，然后设置昵称和密码。\n\n用户问题: 国外手机号可以吗？\n改写后: 国外手机号是否可以注册微信账号\n", "enabl
e_query_expansion": true}
web_search_config   | {"api_key": "", "provider": "duckduckgo", "blacklist": [], "max_results": 5, "include_date"
: true, "compression_method": "none"}


### users
                                    Table "public.users"
         Column         |           Type           | Collation | Nullable |      Default       
------------------------+--------------------------+-----------+----------+--------------------
 id                     | character varying(36)    |           | not null | uuid_generate_v4()
 username               | character varying(100)   |           | not null | 
 email                  | character varying(255)   |           | not null | 
 password_hash          | character varying(255)   |           | not null | 
 avatar                 | character varying(500)   |           |          | 
 tenant_id              | integer                  |           |          | 
 is_active              | boolean                  |           | not null | true
 created_at             | timestamp with time zone |           |          | CURRENT_TIMESTAMP
 updated_at             | timestamp with time zone |           |          | CURRENT_TIMESTAMP
 deleted_at             | timestamp with time zone |           |          | 
 can_access_all_tenants | boolean                  |           | not null | false
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "idx_users_deleted_at" btree (deleted_at)
    "idx_users_email" btree (email)
    "idx_users_tenant_id" btree (tenant_id)
    "idx_users_username" btree (username)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
    "users_username_key" UNIQUE CONSTRAINT, btree (username)
Foreign-key constraints:
    "fk_users_tenant" FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
Referenced by:
    TABLE "auth_tokens" CONSTRAINT "fk_auth_tokens_user" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE


---

-[ RECORD 1 ]----------+-------------------------------------------------------------
id                     | 381b1418-f2d7-4fe2-900d-ae2014faf09f
username               | user1
email                  | user1@qq.com
password_hash          | $2a$10$aUfjwLLTbmMAhxV8J.epNOcruwTUBbAn8LyXBrkSkwEQekNRVOnF2
avatar                 | 
tenant_id              | 10000
is_active              | t
created_at             | 2026-02-08 11:40:19.167834+00
updated_at             | 2026-02-08 11:40:19.167835+00
deleted_at             | 
can_access_all_tenants | f
-[ RECORD 2 ]----------+-------------------------------------------------------------
id                     | 2cb94175-8589-4c4d-968a-ce08beb429c2
username               | user2
email                  | user2@qq.com
password_hash          | $2a$10$DMTFUonDj7tYgZtCvgwchOdYTmaHch7raTzfE9p/kTxGhTFBpPzGC
avatar                 | 
tenant_id              | 10001
is_active              | t
created_at             | 2026-02-08 12:07:45.423011+00
updated_at             | 2026-02-08 12:07:45.423011+00
deleted_at             | 
can_access_all_tenants | f