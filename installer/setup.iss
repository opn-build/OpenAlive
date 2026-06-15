; OpenAlive (Go) — Inno Setup 6 script
; No admin required  |  Installs to %LOCALAPPDATA%\OpenAlive
; The Go build is a single native exe (no runtime, no onedir folder).

#define AppName      "OpenAlive"
#define AppVersion   "1.1.0"
#define AppPublisher "opn-build"
#define AppURL       "https://github.com/opn-build/OpenAlive-Releases"
#define AppExeName   "OpenAlive.exe"

[Setup]
AppId={{6F4E2A1B-8C3D-4E5F-9A7B-2C1D3E4F5A6B}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}
AppUpdatesURL={#AppURL}/releases

DefaultDirName={localappdata}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes

PrivilegesRequired=lowest
MinVersion=10.0

SetupIconFile=..\assets\icon.ico
UninstallDisplayIcon={app}\{#AppExeName}

OutputDir=Output
OutputBaseFilename={#AppName}_Setup_v{#AppVersion}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern

VersionInfoVersion={#AppVersion}
VersionInfoCompany={#AppPublisher}
VersionInfoDescription={#AppName} Setup
VersionInfoProductName={#AppName}
VersionInfoProductVersion={#AppVersion}

[Languages]
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; \
  Description: "{cm:CreateDesktopIcon}"; \
  GroupDescription: "{cm:AdditionalIcons}"; \
  Flags: unchecked

[Files]
; Single native exe produced by: go build -ldflags "-H windowsgui -s -w"
Source: "..\build\OpenAlive.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName}";              Filename: "{app}\{#AppExeName}"
Name: "{group}\Desinstalar {#AppName}";  Filename: "{uninstallexe}"
Name: "{userdesktop}\{#AppName}";        Filename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExeName}"; \
  Description: "Iniciar {#AppName} ahora"; \
  Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "reg.exe"; \
  Parameters: "delete ""HKCU\Software\Microsoft\Windows\CurrentVersion\Run"" /v {#AppName} /f"; \
  Flags: runhidden; RunOnceId: "RemoveAutostart"
