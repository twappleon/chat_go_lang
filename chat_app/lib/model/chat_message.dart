class ChatMessage {
  final String id;
  final String type;
  final String sender;
  final String receiver;
  final String message;
  final DateTime timestamp;

  ChatMessage({
    required this.id,
    required this.type,
    required this.sender,
    required this.receiver,
    required this.message,
    required this.timestamp,
  });

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id'] ?? '',
      type: json['type'] ?? 'chat',
      sender: json['sender'] ?? '',
      receiver: json['receiver'] ?? '',
      message: json['message'] ?? '',
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type,
      'sender': sender,
      'receiver': receiver,
      'message': message,
      'timestamp': timestamp.toIso8601String(),
    };
  }
}