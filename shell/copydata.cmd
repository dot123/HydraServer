@echo off

set source_dir=..\data
set target_dir=..\bin\data

REM ���Ŀ���ļ����Ƿ���ڣ�����������򴴽�
if not exist "%target_dir%" (
    mkdir "%target_dir%"
)

REM ����Դ�ļ��е����ݵ�Ŀ���ļ���
xcopy "%source_dir%\*" "%target_dir%\" /e /y

REM ��鸴�Ʋ����Ƿ�ɹ�
if errorlevel 1 (
    echo �����ļ�ʱ���ִ���
) else (
    echo �ļ����Ƴɹ���
)

pause
