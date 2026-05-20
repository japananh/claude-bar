import Foundation

/// Represents a live Claude process found in ~/.claude/sessions/.
struct RunningSession {
    let pid: Int
    let cwd: String
    let entrypoint: String

    /// Short path label: last 2 components (e.g. "project/src")
    var locationLabel: String {
        let url = URL(fileURLWithPath: cwd)
        let last   = url.lastPathComponent
        let parent = url.deletingLastPathComponent().lastPathComponent
        return parent.isEmpty ? last : "\(parent)/\(last)"
    }

    /// Human-readable origin (Terminal, VSCode, …)
    var typeLabel: String {
        switch entrypoint {
        case "cli":            return "Terminal"
        case "claude-vscode":  return "VSCode"
        case "claude-cursor":  return "Cursor"
        case "claude-windsurf":return "Windsurf"
        default:               return entrypoint
        }
    }

    /// Read all live Claude sessions from ~/.claude/sessions/*.json.
    static func readAll() -> [RunningSession] {
        let dir = URL(fileURLWithPath: NSHomeDirectory()).appendingPathComponent(".claude/sessions")
        guard let entries = try? FileManager.default
            .contentsOfDirectory(at: dir, includingPropertiesForKeys: nil) else { return [] }
        return entries.filter { $0.pathExtension == "json" }.compactMap { url in
            guard let data = try? Data(contentsOf: url),
                  let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
                  let pid  = json["pid"] as? Int,
                  let cwd  = json["cwd"] as? String,
                  let ep   = json["entrypoint"] as? String,
                  kill(Int32(pid), 0) == 0 || errno == EPERM else { return nil }
            return RunningSession(pid: pid, cwd: cwd, entrypoint: ep)
        }
    }
}
