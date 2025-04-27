title judge
set auth_path=C:\StarsProject\starshelper-auth
set meta_path=%auth_path%\meta
set foundation_path=%auth_path%\foundation
set config_path=%auth_path%\judge
judge --module=judge ^
    --meta-config=%config_path%\meta.yaml ^
    --foundation-config=%foundation_path%\foundation.yaml ^
    --config=%config_path%\config.yaml ^
    --log-config=%meta_path%\log.yaml
