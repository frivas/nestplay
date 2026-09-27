import AVKit
import UIKit

@MainActor
final class LibraryPlayer: NSObject {
  static let shared = LibraryPlayer()

  private var urls: [URL] = []
  private var index = 0
  private var endObserver: NSObjectProtocol?

  func present(urls: [URL], start: Int) {
    guard urls.indices.contains(start), let presenter = Self.topViewController() else { return }
    self.urls = urls
    self.index = start
    let controller = AVPlayerViewController()
    presenter.present(controller, animated: true) { [weak self] in
      self?.playCurrent(on: controller)
    }
  }

  private func playCurrent(on controller: AVPlayerViewController) {
    clearEndObserver()
    let player = AVPlayer(url: urls[index])
    controller.player = player
    controller.skippingBehavior = .default
    controller.isSkipBackwardEnabled = true
    controller.isSkipForwardEnabled = true
    applyEpisodeActions(on: controller)
    if let item = player.currentItem {
      endObserver = NotificationCenter.default.addObserver(
        forName: .AVPlayerItemDidPlayToEndTime,
        object: item,
        queue: .main
      ) { [weak self, weak controller] _ in
        guard let controller else { return }
        Task { @MainActor in
          self?.playNextAfterEnd(on: controller)
        }
      }
    }
    player.play()
  }

  private func applyEpisodeActions(on controller: AVPlayerViewController) {
    let previous = UIAction(
      title: "Previous",
      image: UIImage(systemName: "backward.end.fill"),
      attributes: index > 0 ? [] : [.disabled]
    ) { [weak self, weak controller] _ in
      guard let controller else { return }
      Task { @MainActor in
        self?.playOffset(-1, on: controller)
      }
    }
    let next = UIAction(
      title: "Next",
      image: UIImage(systemName: "forward.end.fill"),
      attributes: index + 1 < urls.count ? [] : [.disabled]
    ) { [weak self, weak controller] _ in
      guard let controller else { return }
      Task { @MainActor in
        self?.playOffset(1, on: controller)
      }
    }
    controller.transportBarCustomMenuItems = [previous, next]
  }

  private func playOffset(_ delta: Int, on controller: AVPlayerViewController) {
    let nextIndex = index + delta
    guard urls.indices.contains(nextIndex) else { return }
    index = nextIndex
    playCurrent(on: controller)
  }

  private func playNextAfterEnd(on controller: AVPlayerViewController) {
    guard controller.presentingViewController != nil else { return }
    playOffset(1, on: controller)
  }

  private func clearEndObserver() {
    if let endObserver {
      NotificationCenter.default.removeObserver(endObserver)
      self.endObserver = nil
    }
  }

  private static func topViewController() -> UIViewController? {
    let scenes = UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }
    let window =
      scenes.flatMap(\.windows).first(where: \.isKeyWindow) ?? scenes.first?.windows.first
    var current = window?.rootViewController
    while let next = current?.presentedViewController { current = next }
    return current
  }
}
