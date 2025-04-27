title web
set auth_path=C:\StarsProject\starsweb-auth
set meta_path=%auth_path%\meta
set foundation_path=%auth_path%\foundation
set config_path=%auth_path%\web
web --module=web ^
    --meta-config=%config_path%\meta.yaml ^
    --foundation-config=%foundation_path%\foundation.yaml ^
    --config=%config_path%\config.yaml ^
    --log-config=%meta_path%\log.yaml
