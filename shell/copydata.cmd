@echo off

set source_dir=..\data
set target_dir=..\bin\data

REM 检查目标文件夹是否存在，如果不存在则创建
if not exist "%target_dir%" (
    mkdir "%target_dir%"
)

REM 复制源文件夹的内容到目标文件夹
xcopy "%source_dir%\*" "%target_dir%\" /e /y

REM 检查复制操作是否成功
if errorlevel 1 (
    echo 复制文件时出现错误。
) else (
    echo 文件复制成功。
)

pause
