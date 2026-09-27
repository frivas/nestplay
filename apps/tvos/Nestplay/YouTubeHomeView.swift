import HealthClient
import SwiftUI
import UIKit

struct YouTubeHomeView: View {
  let server: String
  @State private var homePage = YouTubePage(state: "checking", message: nil, items: [], cursor: nil)
  @State private var searchPage = YouTubePage(
    state: "connected", message: nil, items: [], cursor: nil)
  @State private var searching = false
  @State private var didSearch = false
  @State private var query = ""
  @State private var homeLoading = false
  @State private var searchLoading = false
  @FocusState private var chromeFocus: ChromeFocus?

  private enum ChromeFocus: Hashable {
    case refresh
    case search
    case close
    case find
  }

  private var page: YouTubePage { searching ? searchPage : homePage }
  private var loading: Bool { searching ? searchLoading : homeLoading }

  var body: some View {
    VStack(alignment: .leading, spacing: 24) {
      chrome
      if loading && page.items.isEmpty {
        ProgressView().tint(NestEgg.cream)
      }
      if searching && !didSearch && searchPage.items.isEmpty && !loading {
        Text("Type a search, then press Find.")
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
      } else if let message = page.message, page.items.isEmpty {
        Text(message)
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
          .accessibilityIdentifier("youtubeHomeEmpty")
      } else if page.items.isEmpty && !loading {
        Text(searching ? "No landscape search results." : "No landscape recommendations right now.")
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
          .accessibilityIdentifier("youtubeHomeEmpty")
      }
      if !page.items.isEmpty {
        PosterGrid(
          items: page.items,
          server: server,
          catalogID: "youtubeHomeCatalog",
          title: { $0.title },
          artwork: { $0.artwork },
          onMoveUpFromTopRow: {
            chromeFocus = searching ? .close : .refresh
          }
        ) {
          if let cursor = page.cursor, !cursor.isEmpty {
            NestChromeButton(
              title: searching ? "More results" : "More recommendations",
              kind: .more
            ) {
              Task { await loadMore() }
            }
          }
        }
      }
    }
    .padding(.horizontal, 72)
    .padding(.top, 28)
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    .accessibilityIdentifier("youtubeHomeSurface")
    .task { await loadShelf() }
    .modifier(SearchExitModifier(enabled: searching, close: closeSearch))
  }

  private var chrome: some View {
    VStack(alignment: .leading, spacing: 20) {
      header
        .focusSection()
      if searching {
        searchField
          .focusSection()
      }
    }
  }

  private var header: some View {
    HStack(alignment: .center, spacing: 24) {
      Text(searching ? "YouTube Search" : "YouTube Home")
        .font(.system(size: 44, weight: .semibold))
        .foregroundStyle(NestEgg.cream)
      Spacer()
      if searching {
        NestChromeButton(title: "Close Search", kind: .close) {
          closeSearch()
        }
        .focused($chromeFocus, equals: .close)
      } else {
        NestChromeButton(
          title: "Refresh",
          kind: .refresh,
          enabled: !homeLoading,
          accessibilityID: "youtubeHomeRefresh"
        ) {
          Task { await refresh() }
        }
        .focused($chromeFocus, equals: .refresh)
        NestChromeButton(
          title: "Search",
          kind: .search,
          enabled: homePage.state != "reconnect_required",
          accessibilityID: "youtubeHomeSearch"
        ) {
          openSearch()
        }
        .focused($chromeFocus, equals: .search)
      }
    }
  }

  private var searchField: some View {
    HStack(alignment: .center, spacing: 24) {
      ActivatingSearchField(text: $query, keyboardActive: !didSearch) {
        Task { await search() }
      }
      .frame(width: TVSearchFieldMetrics.width, height: TVSearchFieldMetrics.height)
      NestChromeButton(
        title: "Find",
        kind: .find,
        enabled: !query.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !searchLoading,
        accessibilityID: "youtubeSearchSubmit"
      ) {
        Task { await search() }
      }
      .focused($chromeFocus, equals: .find)
    }
  }

  private func openSearch() {
    searching = true
    didSearch = false
    searchPage = YouTubePage(state: "connected", message: nil, items: [], cursor: nil)
  }

  private func closeSearch() {
    searching = false
    didSearch = false
    query = ""
    searchPage = YouTubePage(state: "connected", message: nil, items: [], cursor: nil)
    if homePage.items.isEmpty {
      Task { await loadShelf() }
    }
  }

  private func loadShelf() async {
    await fetchHome(refresh: false)
  }

  private func refresh() async {
    await fetchHome(refresh: true)
  }

  private func fetchHome(refresh: Bool) async {
    homeLoading = true
    defer { homeLoading = false }
    do {
      homePage = try await HealthClient().youtubeHome(address: server, refresh: refresh)
    } catch {
      homePage = YouTubePage(
        state: "unavailable",
        message: "YouTube Home could not be loaded.",
        items: [],
        cursor: nil)
    }
  }

  private func search() async {
    let q = query.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !q.isEmpty else { return }
    didSearch = true
    searchLoading = true
    defer { searchLoading = false }
    do {
      searchPage = try await HealthClient().youtubeSearch(address: server, query: q)
    } catch {
      searchPage = YouTubePage(
        state: "unavailable",
        message: "YouTube Search could not be loaded.",
        items: [],
        cursor: nil)
    }
  }

  private func loadMore() async {
    guard let cursor = page.cursor, !cursor.isEmpty else { return }
    if searching {
      searchLoading = true
    } else {
      homeLoading = true
    }
    defer {
      if searching {
        searchLoading = false
      } else {
        homeLoading = false
      }
    }
    do {
      let next: YouTubePage
      if searching {
        next = try await HealthClient().youtubeSearch(
          address: server, query: query, cursor: cursor)
        searchPage = YouTubePage(
          state: next.state,
          message: next.message,
          items: searchPage.items + next.items,
          cursor: next.cursor)
      } else {
        next = try await HealthClient().youtubeHome(address: server, cursor: cursor)
        homePage = YouTubePage(
          state: next.state,
          message: next.message,
          items: homePage.items + next.items,
          cursor: next.cursor)
      }
    } catch {
      let failed = YouTubePage(
        state: page.state,
        message: "Could not load more videos.",
        items: page.items,
        cursor: nil)
      if searching {
        searchPage = failed
      } else {
        homePage = failed
      }
    }
  }
}

private enum TVSearchFieldMetrics {
  /// One-line tvOS control height, matching unfocused buttons.
  static let height: CGFloat = 66
  /// ~40 characters at headline size on 1080p; enough for a query, not a paragraph.
  static let width: CGFloat = 780
}

private struct ActivatingSearchField: UIViewRepresentable {
  @Binding var text: String
  var keyboardActive: Bool
  var onSubmit: () -> Void

  func makeCoordinator() -> Coordinator {
    Coordinator(text: $text, onSubmit: onSubmit)
  }

  func makeUIView(context: Context) -> CompactSearchField {
    let field = CompactSearchField()
    field.placeholder = "Search YouTube"
    field.returnKeyType = .search
    field.autocorrectionType = .no
    field.autocapitalizationType = .none
    field.spellCheckingType = .no
    field.keyboardAppearance = .dark
    field.font = UIFont.preferredFont(forTextStyle: .headline)
    field.borderStyle = .roundedRect
    field.accessibilityIdentifier = "youtubeSearchField"
    field.delegate = context.coordinator
    field.addTarget(
      context.coordinator, action: #selector(Coordinator.editingChanged(_:)), for: .editingChanged)
    field.setContentHuggingPriority(.required, for: .vertical)
    field.setContentCompressionResistancePriority(.required, for: .vertical)
    return field
  }

  func updateUIView(_ field: CompactSearchField, context: Context) {
    context.coordinator.onSubmit = onSubmit
    if field.text != text {
      field.text = text
    }
    if !keyboardActive, field.isFirstResponder {
      field.resignFirstResponder()
    }
  }

  func sizeThatFits(_ proposal: ProposedViewSize, uiView: CompactSearchField, context: Context)
    -> CGSize?
  {
    CGSize(width: TVSearchFieldMetrics.width, height: TVSearchFieldMetrics.height)
  }

  final class Coordinator: NSObject, UITextFieldDelegate {
    var text: Binding<String>
    var onSubmit: () -> Void

    init(text: Binding<String>, onSubmit: @escaping () -> Void) {
      self.text = text
      self.onSubmit = onSubmit
    }

    @objc func editingChanged(_ sender: UITextField) {
      let value = sender.text ?? ""
      if text.wrappedValue != value {
        text.wrappedValue = value
      }
    }

    func textFieldShouldReturn(_ textField: UITextField) -> Bool {
      onSubmit()
      return true
    }
  }
}

private final class CompactSearchField: UITextField {
  private var didActivateKeyboard = false

  override var intrinsicContentSize: CGSize {
    CGSize(width: TVSearchFieldMetrics.width, height: TVSearchFieldMetrics.height)
  }

  override func didMoveToWindow() {
    super.didMoveToWindow()
    guard window != nil, !didActivateKeyboard else { return }
    didActivateKeyboard = true
    _ = becomeFirstResponder()
    Task { @MainActor in
      await TVSearchKeyboard.capitalizeSearchSubmitTitle()
    }
  }
}

@MainActor
private enum TVSearchKeyboard {
  static func capitalizeSearchSubmitTitle() async {
    for _ in 0..<6 {
      applySearchTitle()
      try? await Task.sleep(nanoseconds: 100_000_000)
    }
  }

  private static func applySearchTitle() {
    for scene in UIApplication.shared.connectedScenes {
      guard let windowScene = scene as? UIWindowScene else { continue }
      for window in windowScene.windows {
        retitleSearch(in: window)
      }
    }
  }

  private static func retitleSearch(in view: UIView) {
    if let button = view as? UIButton {
      if var config = button.configuration,
        let title = config.title,
        title.compare("search", options: .caseInsensitive) == .orderedSame
      {
        config.title = "Search"
        button.configuration = config
      }
      for state: UIControl.State in [.normal, .focused, .highlighted, .selected, .disabled] {
        if let title = button.title(for: state),
          title.compare("search", options: .caseInsensitive) == .orderedSame
        {
          button.setTitle("Search", for: state)
        }
      }
    }
    if let label = view as? UILabel,
      let text = label.text,
      text.compare("search", options: .caseInsensitive) == .orderedSame
    {
      label.text = "Search"
    }
    for subview in view.subviews {
      retitleSearch(in: subview)
    }
  }
}

private struct SearchExitModifier: ViewModifier {
  let enabled: Bool
  let close: () -> Void

  func body(content: Content) -> some View {
    if enabled {
      content.onExitCommand(perform: close)
    } else {
      content
    }
  }
}
